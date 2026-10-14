require 'digest'
require 'fileutils'
require 'json'
require 'pathname'
require 'yaml'

# Shared by the runner and the server. It never loads business credentials.
module Release
  ROOT = File.expand_path('..', __dir__)
  TAG = /\A(back|admin|app)-(test-)?v([1-9]\d*|\d+\.\d+\.\d+)\z/
  ENVIRONMENTS = {
    'test' => {
      'root' => '/www/wwwroot/Leave_test', 'database' => 'leave_list_test',
      'config' => 'config.test.yaml', 'project' => 'Leave_test',
      'admin_host' => 'admin-test.leave.mhspace.dpdns.org', 'admin_dist' => 'dist-test',
      'backend_dir' => '/www/wwwroot/Leave_test/backend', 'process_user' => 'www',
      'admin_port' => 80
    },
    'prod' => {
      'root' => '/www/wwwroot/Leave', 'database' => 'leave_list',
      'config' => 'config.server.yaml', 'project' => 'Leave',
      'admin_host' => 'admin.leave.mhspace.dpdns.org', 'admin_dist' => 'dist',
      'backend_dir' => '/www/wwwroot/Leave/backend', 'process_user' => 'www',
      'admin_port' => 8001
    }
  }.freeze

  def self.environment(info)
    match = TAG.match(info.fetch('tag'))
    raise 'Invalid release tag' unless match
    name = match[2] ? 'test' : 'prod'
    raise 'Release environment does not match tag' if info.key?('environment') && info['environment'] != name
    name
  end

  def self.server_config(path, environment)
    profile = ENVIRONMENTS.fetch(environment)
    config = YAML.safe_load(File.read(path), aliases: false)
    mysql = config.fetch('mysql')
    raise "Refusing database outside #{environment}: expected #{profile['database']}" unless mysql.fetch('database') == profile['database']
    raise 'MySQL must use loopback' unless %w[127.0.0.1 localhost ::1].include?(mysql.fetch('host'))
    root = profile.fetch('root')
    raise 'Unexpected APK storage path' unless config.fetch('app').fetch('package_dir') == "#{root}/packages/android"
    raise 'Unexpected uploads path' unless config.fetch('storage').fetch('root_dir') == "#{root}/uploads"
    port = integer(config.fetch('app').fetch('port'), 'app.port')
    raise 'Invalid app.port' unless port <= 65_535
    config
  end

  def self.integer(value, name)
    raise "#{name} must be a positive integer" unless value.is_a?(Integer) && value.positive?
    value
  end

  def self.read(tag, root = ROOT)
    match = TAG.match(tag.to_s)
    raise 'Only back/admin/app release tags are supported' unless match
    mode, test_prefix, version = match.captures
    config = YAML.safe_load(File.read(File.join(root, 'deploy/release.yaml')), aliases: false)
    info = {
      'tag' => tag,
      'mode' => mode,
      'environment' => test_prefix ? 'test' : 'prod',
      'backend_version' => integer(config.fetch('backend').fetch('version'), 'backend.version')
    }
    if mode == 'app'
      raise 'App tag requires X.Y.Z' unless version.match?(/\A\d+\.\d+\.\d+\z/)
      manifest = File.read(File.join(root, 'frontend/manifest.json'))
      name = manifest[/"versionName"\s*:\s*"([^"]+)"/, 1]
      code = manifest[/"versionCode"\s*:\s*"(\d+)"/, 1]
      raise "#{tag}（标签版本 #{version}）与 frontend/manifest.json 的 versionName=#{name.inspect} 不一致，发布已停止" unless name == version
      raise 'Invalid manifest versionCode' unless code && code.to_i.positive? && code.to_i <= 4_294_967_295
      android = config.fetch('app').fetch('android')
      notes = android.fetch('release_notes')
      raise 'release_notes must be a string list' unless notes.is_a?(Array) && notes.all? { |n| n.is_a?(String) }
      directory = android.fetch('apk_directory')
      raise 'apk_directory must be repository-relative' unless directory.is_a?(String) && !directory.empty? && !Pathname.new(directory).absolute?
      path = File.realpath(File.join(root, directory))
      raise 'APK directory escapes repository' unless path.start_with?(File.realpath(root) + File::SEPARATOR)
      apks = Dir.children(path).select { |n| n.end_with?('.apk') && File.file?(File.join(path, n)) }
      raise 'APK directory must contain exactly one .apk file' unless apks.length == 1
      apk = File.join(path, apks.first)
      raise 'APK must not be a symlink' if File.symlink?(apk)
      raise 'APK path contains control characters' if apk.match?(/[\r\n\x00]/)
      info['app'] = {
        'version_name' => name,
        'version_code' => code.to_i,
        'apk_name' => "leave-android-#{code.to_i}-#{name}.apk",
        'apk_path' => apk,
        'release_notes' => notes
      }
    else
      raise 'Backend/admin tag requires a positive integer' unless version.match?(/\A[1-9]\d*\z/)
      if mode == 'admin'
        info['admin_version'] = integer(config.fetch('admin').fetch('version'), 'admin.version')
        raise "#{tag}（标签版本 #{version}）与 deploy/release.yaml 的 admin.version=#{info['admin_version']} 不一致，发布已停止" unless info['admin_version'].to_s == version
      else
        raise "#{tag}（标签版本 #{version}）与 deploy/release.yaml 的 backend.version=#{info['backend_version']} 不一致，发布已停止" unless info['backend_version'].to_s == version
      end
    end
    info
  end

  def self.validate(info)
    match = TAG.match(info.fetch('tag'))
    raise 'Invalid release tag' unless match && match[1] == info.fetch('mode')
    mode, _, version = match.captures
    environment(info)
    backend = integer(info.fetch('backend_version'), 'backend_version')
    case mode
    when 'back'
      raise 'Invalid backend version' unless version == backend.to_s
    when 'admin'
      raise 'Invalid admin version' unless version == integer(info.fetch('admin_version'), 'admin_version').to_s
    when 'app'
      app = info.fetch('app')
      code = integer(app.fetch('version_code'), 'version_code')
      raise 'Invalid app version' unless version.match?(/\A\d+\.\d+\.\d+\z/) && version == app.fetch('version_name')
      raise 'Invalid APK filename' unless app.fetch('apk_name') == "leave-android-#{code}-#{version}.apk"
      notes = app.fetch('release_notes')
      raise 'Invalid release notes' unless notes.is_a?(Array) && notes.all? { |n| n.is_a?(String) }
    end
    info
  end

  def self.bundle(info, destination, root = ROOT)
    validate(info)
    raise 'Bundle destination must be empty' if File.exist?(destination) && !Dir.empty?(destination)
    FileUtils.mkdir_p(destination)
    { 'bin/Leave' => 'backend/Leave', 'bin/migrate' => 'scripts/migrate' }.each do |source, target|
      FileUtils.mkdir_p(File.dirname(File.join(destination, target)))
      FileUtils.cp(File.join(root, source), File.join(destination, target))
    end
    FileUtils.cp_r(File.join(root, 'backend/database/migrations'), destination)
    FileUtils.mkdir_p(File.join(destination, 'scripts'))
    %w[release.rb deploy-release.sh switch-and-restart.sh run-migrations.sh publish-version.sh].each do |name|
      FileUtils.cp(File.join(root, 'deploy', name), File.join(destination, 'scripts', name))
    end
    if info['mode'] == 'admin'
      dist = ENVIRONMENTS.fetch(environment(info)).fetch('admin_dist')
      FileUtils.cp_r(File.join(root, 'admin', dist), File.join(destination, 'admin'))
    elsif info['mode'] == 'app'
      FileUtils.mkdir_p(File.join(destination, 'app'))
      FileUtils.cp(info.fetch('app').fetch('apk_path'), File.join(destination, 'app', info['app']['apk_name']))
      File.write(File.join(destination, 'app/app-version.sql'), version_sql(info))
    end
    metadata = Marshal.load(Marshal.dump(info))
    metadata['app']&.delete('apk_path')
    File.write(File.join(destination, 'release.json'), JSON.pretty_generate(metadata) + "\n")
    files = Dir.glob(File.join(destination, '**/*'), File::FNM_DOTMATCH).select { |p| File.file?(p) }.sort
    sums = files.map do |path|
      relative = Pathname.new(path).relative_path_from(Pathname.new(destination)).to_s
      raise 'Unsupported bundle filename' unless relative.match?(/\A[A-Za-z0-9_.\/-]+\z/)
      "#{Digest::SHA256.file(path).hexdigest}  #{relative}\n"
    end
    File.write(File.join(destination, 'SHA256SUMS'), sums.join)
    verify_bundle(destination)
  end

  def self.verify_bundle(directory)
    root = File.realpath(directory)
    actual = Dir.glob(File.join(root, '**/*'), File::FNM_DOTMATCH).reject { |p| ['.', '..'].include?(File.basename(p)) }
    raise 'Bundle symlinks are not allowed' if actual.any? { |p| File.symlink?(p) }
    expected = File.readlines(File.join(root, 'SHA256SUMS')).map do |line|
      match = /\A([0-9a-f]{64})  ([A-Za-z0-9_.\/-]+)\n?\z/.match(line)
      raise 'Invalid checksum entry' unless match
      relative = match[2]
      path = File.expand_path(relative, root)
      raise 'Checksum path escapes bundle' unless path.start_with?(root + File::SEPARATOR)
      raise "Checksum mismatch: #{relative}" unless File.file?(path) && Digest::SHA256.file(path).hexdigest == match[1]
      path
    end
    files = actual.select { |p| File.file?(p) && p != File.join(root, 'SHA256SUMS') }
    raise 'Unexpected or duplicate bundle files' unless files.sort == expected.sort && expected.uniq.length == expected.length
    info = validate(JSON.parse(File.read(File.join(root, 'release.json'))))
    required = %w[backend/Leave scripts/migrate scripts/release.rb scripts/deploy-release.sh scripts/switch-and-restart.sh scripts/run-migrations.sh scripts/publish-version.sh]
    required += ['admin/index.html'] if info['mode'] == 'admin'
    required += ["app/#{info['app']['apk_name']}", 'app/app-version.sql'] if info['mode'] == 'app'
    required.each { |p| raise "Missing bundle file: #{p}" unless File.file?(File.join(root, p)) }
    raise 'No migrations in bundle' if Dir.glob(File.join(root, 'migrations/*.up.sql')).empty?
    info
  end

  # Compare component bytes, not whole-repository commits or release tag metadata.
  def self.check_version(kind, bundle, target)
    return unless File.exist?(target) || File.symlink?(target)
    raise "Existing version is a symlink: #{target}" if File.symlink?(target)
    info = JSON.parse(File.read(File.join(bundle, 'release.json')))
    if kind == 'app'
      source = File.join(bundle, 'app', info.fetch('app').fetch('apk_name'))
      raise "Existing APK differs or is incomplete: #{target}" unless File.file?(target) && Digest::SHA256.file(source).hexdigest == Digest::SHA256.file(target).hexdigest
      return
    end
    raise "Existing version is not a directory: #{target}" unless File.directory?(target)
    paths = Dir.glob(File.join(target, '**/*'), File::FNM_DOTMATCH).reject { |p| %w[. ..].include?(File.basename(p)) }
    raise "Existing version contains symlinks: #{target}" if paths.any? { |p| File.symlink?(p) }
    expected = if kind == 'backend'
      metadata = File.join(target, 'release.json')
      raise "Existing backend version is incomplete: #{target}" unless File.file?(metadata)
      previous = validate(JSON.parse(File.read(metadata)))
      raise 'Existing backend metadata has a different environment' unless environment(previous) == environment(info)
      raise 'Existing backend metadata has a different version' unless previous.fetch('backend_version') == info.fetch('backend_version')
      { 'Leave' => File.join(bundle, 'backend/Leave') }
    elsif kind == 'admin'
      source = File.join(bundle, 'admin')
      Dir.glob(File.join(source, '**/*'), File::FNM_DOTMATCH).select { |p| File.file?(p) }.to_h do |p|
        [Pathname.new(p).relative_path_from(Pathname.new(source)).to_s, p]
      end
    else
      raise 'Unknown version kind'
    end
    files = paths.select { |p| File.file?(p) }.map { |p| Pathname.new(p).relative_path_from(Pathname.new(target)).to_s }
    required = expected.keys + (kind == 'backend' ? ['release.json'] : [])
    raise "Existing #{kind} version differs or is incomplete: #{target}" unless files.sort == required.sort
    expected.each do |relative, source|
      raise "Existing #{kind} artifact differs: #{relative}; increment the version" unless Digest::SHA256.file(source).hexdigest == Digest::SHA256.file(File.join(target, relative)).hexdigest
    end
  end

  def self.version_sql(info)
    validate(info)
    raise 'Only App releases produce version SQL' unless info['mode'] == 'app'
    app = info.fetch('app')
    # Hex literals work independently of the server's string-escaping SQL mode.
    literal = ->(s) { "CONVERT(0x#{s.encode('UTF-8').unpack1('H*')} USING utf8mb4)" }
    name = literal.call(app.fetch('version_name'))
    package = literal.call(app.fetch('apk_name'))
    notes = literal.call(JSON.generate(app.fetch('release_notes')))
    <<~SQL
      START TRANSACTION;
      UPDATE app_versions SET status = 'archived'
      WHERE platform = 'android' AND status = 'published' AND version_code <> #{app.fetch('version_code')};
      INSERT INTO app_versions (platform, version_code, version_name, package_file, release_notes, status)
      VALUES ('android', #{app.fetch('version_code')}, #{name}, #{package}, #{notes}, 'published')
      ON DUPLICATE KEY UPDATE
        version_name = #{name},
        package_file = #{package},
        release_notes = #{notes},
        status = 'published',
        updated_at = CURRENT_TIMESTAMP;
      COMMIT;
    SQL
  end
end

if $PROGRAM_NAME == __FILE__
  begin
    case ARGV.shift
    when 'verify'
      info = Release.verify_bundle(ARGV.fetch(0))
      puts info.fetch('mode'), info.fetch('backend_version'), info.fetch('admin_version', ''), info.dig('app', 'apk_name').to_s, Release.environment(info)
    when 'environment'
      profile = Release::ENVIRONMENTS.fetch(ARGV.fetch(0))
      puts profile.fetch('root'), profile.fetch('config'), profile.fetch('project'), profile.fetch('admin_host'), profile.fetch('backend_dir'), profile.fetch('process_user'), profile.fetch('admin_port')
    when 'server-config'
      config = Release.server_config(ARGV.fetch(0), ARGV.fetch(1, 'test'))
      puts config.fetch('app').fetch('port')
    when 'database-config'
      config = Release.server_config(ARGV.fetch(0), ARGV.fetch(1))
      %w[host port database user].each { |key| puts config.fetch('mysql').fetch(key) }
    when 'check-version'
      Release.check_version(ARGV.fetch(0), ARGV.fetch(1), ARGV.fetch(2))
    else
      abort 'Usage: release.rb verify BUNDLE | environment test|prod | server-config YAML ENV | database-config YAML ENV | check-version KIND BUNDLE TARGET'
    end
  rescue StandardError => e
    warn e.message
    exit 1
  end
end
