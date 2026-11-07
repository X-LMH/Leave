#!/usr/bin/env ruby
require 'open3'
require 'tmpdir'
require_relative 'release'

# All external arguments are passed as argv, never interpreted as shell source.
module GitHubRelease
  def self.run(*args)
    output, error, status = Open3.capture3(*args)
    raise "#{args.first} failed: #{error}" unless status.success?
    output
  end

  def self.version(tag)
    match = /\Aapp-v(\d+)\.(\d+)\.(\d+)\z/.match(tag)
    match && match.captures.map(&:to_i)
  end

  def self.previous_tag(tag)
    current = version(tag)
    tags = run('git', 'tag', '--merged', 'HEAD').lines.map(&:strip).select { |t| version(t) && (version(t) <=> current) == -1 }
    commits = tags.group_by { |t| run('git', 'rev-list', '-1', t).strip }
    run('git', 'rev-list', 'HEAD').each_line do |sha|
      candidates = commits[sha.strip]
      return candidates.max_by { |t| version(t) } if candidates
    end
    nil
  end

  def self.markdown(info, repo, previous)
    app = info.fetch('app')
    notes = Release.normalize_notes(app.fetch('release_notes'))
    sections = if notes.is_a?(Array)
      [['更新内容', notes]]
    else
      Release::NOTE_LABELS.map { |key, label| [label, notes.fetch(key)] }
    end
    body = sections.filter_map { |label, items| "### #{label}\n\n#{items.map { |n| "- #{n}" }.join("\n")}\n" unless items.empty? }.join("\n")
    body = "该版本暂无更新说明。\n" if body.empty?
    base = "https://github.com/#{repo}"
    body += "\n### 下载\n\n| 平台 | 下载 |\n| --- | --- |\n| Android | [Leave #{app['version_name']} APK](#{base}/releases/download/#{info['tag']}/#{app['apk_name']}) |\n"
    changes = previous ? "#{base}/compare/#{previous}...#{info['tag']}" : "#{base}/commits/#{info['tag']}"
    body + "\n**完整更新记录**：[查看提交](#{changes})\n"
  end

  def self.prepare(tag, repo, directory)
    raise 'Only production App releases are supported' unless version(tag)
    raise 'Invalid repository' unless repo.match?(%r{\A[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+\z})
    info = Release.read(tag)
    FileUtils.mkdir_p(directory)
    app = info.fetch('app')
    target = File.join(directory, app.fetch('apk_name'))
    FileUtils.cp(app.fetch('apk_path'), target)
    File.write(File.join(directory, 'SHA256SUMS'), "#{Digest::SHA256.file(target).hexdigest}  #{app['apk_name']}\n")
    File.write(File.join(directory, 'notes.md'), markdown(info, repo, previous_tag(tag)))
    File.write(File.join(directory, 'metadata.json'), JSON.generate({ 'tag' => tag, 'title' => "v#{app['version_name']}", 'apk_name' => app['apk_name'] }))
  end

  def self.publish(directory, repo, tag)
    metadata = JSON.parse(File.read(File.join(directory, 'metadata.json')))
    raise 'Release tag mismatch' unless version(tag) && metadata.fetch('tag') == tag
    apk = metadata.fetch('apk_name')
    raise 'Invalid APK filename' unless apk.match?(/\Aleave-android-\d+-\d+\.\d+\.\d+\.apk\z/)
    expected = "#{Digest::SHA256.file(File.join(directory, apk)).hexdigest}  #{apk}\n"
    raise 'APK checksum mismatch' unless File.read(File.join(directory, 'SHA256SUMS')) == expected
    # Listing releases distinguishes an absent release from auth/network failures.
    releases = JSON.parse(run('gh', 'api', '--paginate', '--slurp', "repos/#{repo}/releases?per_page=100")).flatten
    existing = releases.find { |r| r.fetch('tag_name') == tag }
    unless existing
      run('gh', 'release', 'create', tag, '--repo', repo, '--verify-tag', '--draft', '--title', metadata.fetch('title'), '--notes-file', File.join(directory, 'notes.md'))
    end
    state = JSON.parse(run('gh', 'release', 'view', tag, '--repo', repo, '--json', 'isDraft,assets,body'))
    raise 'Published release notes differ' if !state.fetch('isDraft') && state.fetch('body') != File.read(File.join(directory, 'notes.md'))
    [apk, 'SHA256SUMS'].each do |name|
      asset = state.fetch('assets').find { |a| a.fetch('name') == name }
      if asset
        Dir.mktmpdir('leave-release-') do |temp|
          run('gh', 'release', 'download', tag, '--repo', repo, '--pattern', name, '--dir', temp)
          raise "Published asset differs: #{name}" unless Digest::SHA256.file(File.join(temp, name)).hexdigest == Digest::SHA256.file(File.join(directory, name)).hexdigest
        end
      else
        raise "Published release missing asset: #{name}" unless state.fetch('isDraft')
        run('gh', 'release', 'upload', tag, File.join(directory, name), '--repo', repo)
        # Verify the uploaded bytes before exposing the release.
        Dir.mktmpdir('leave-release-') do |temp|
          run('gh', 'release', 'download', tag, '--repo', repo, '--pattern', name, '--dir', temp)
          raise "Uploaded asset differs: #{name}" unless Digest::SHA256.file(File.join(temp, name)).hexdigest == Digest::SHA256.file(File.join(directory, name)).hexdigest
        end
      end
    end
    newer = releases.any? { |r| !r['draft'] && !r['prerelease'] && version(r['tag_name']) && (version(r['tag_name']) <=> version(tag)) == 1 }
    if state.fetch('isDraft')
      run('gh', 'release', 'edit', tag, '--repo', repo, '--title', metadata.fetch('title'), '--notes-file', File.join(directory, 'notes.md'), '--draft=false', "--latest=#{!newer}")
    end
    "https://github.com/#{repo}/releases/tag/#{tag}"
  end
end

if $PROGRAM_NAME == __FILE__
  begin
    command, directory = ARGV
    repo = ENV.fetch('GITHUB_REPOSITORY')
    tag = ENV.fetch('RELEASE_TAG')
    case command
    when 'prepare' then GitHubRelease.prepare(tag, repo, directory)
    when 'publish' then puts GitHubRelease.publish(directory, repo, tag)
    else raise 'Usage: github-release.rb prepare|publish DIRECTORY'
    end
  rescue StandardError => error
    warn error.message
    if ENV['GITHUB_STEP_SUMMARY']
      File.open(ENV['GITHUB_STEP_SUMMARY'], 'a') { |f| f.puts "GitHub Release 未完成：#{error.message}。已成功的部署不会回滚；修复后仅重跑 Release job。" }
    end
    exit 1
  end
end
