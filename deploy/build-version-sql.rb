#!/usr/bin/env ruby

require 'json'
require 'optparse'
require 'yaml'

options = { manifest: File.expand_path('../frontend/manifest.json', __dir__), release: File.expand_path('releases/android.yaml', __dir__), platform: 'android' }
OptionParser.new do |parser|
  parser.on('--manifest PATH') { |path| options[:manifest] = path }
  parser.on('--release PATH') { |path| options[:release] = path }
  parser.on('--platform PLATFORM') { |platform| options[:platform] = platform }
end.parse!

manifest = File.read(options[:manifest])
read_manifest_value = lambda do |name|
  match = manifest.match(/"#{Regexp.escape(name)}"\s*:\s*"([^"]+)"/)
  abort "Missing #{name} in manifest" unless match && !match[1].strip.empty?
  match[1].strip
end
version_name = read_manifest_value.call('versionName')
version_code = read_manifest_value.call('versionCode')
abort "Invalid versionCode: #{version_code.inspect}" unless version_code.match?(/\A\d+\z/) && version_code.to_i.positive?
release = YAML.safe_load(File.read(options[:release]), permitted_classes: [], aliases: false) || {}
release_notes = release.fetch('release_notes', [])
unless release_notes.is_a?(Array) && release_notes.all? { |note| note.is_a?(String) }
  abort 'release_notes must be a YAML list of strings'
end
info = { 'platform' => options[:platform], 'version_code' => version_code.to_i, 'version_name' => version_name, 'package_file' => "leave-#{options[:platform]}-#{version_code}-#{version_name}.apk", 'release_notes' => release_notes }
quote = ->(value) { "'#{value.to_s.gsub("'", "''")}'" }
notes = JSON.generate(info.fetch('release_notes'))

puts <<~SQL
  START TRANSACTION;
  UPDATE `app_versions`
  SET `status` = 'archived'
  WHERE `platform` = #{quote.call(info.fetch('platform'))}
    AND `status` = 'published';
  INSERT INTO `app_versions`
    (`platform`, `version_code`, `version_name`, `package_file`, `release_notes`, `status`)
  VALUES (
    #{quote.call(info.fetch('platform'))},
    #{Integer(info.fetch('version_code'))},
    #{quote.call(info.fetch('version_name'))},
    #{quote.call(info.fetch('package_file'))},
    #{quote.call(notes)},
    'published'
  );
  COMMIT;
SQL
