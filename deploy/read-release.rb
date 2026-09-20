#!/usr/bin/env ruby

require 'json'
require 'optparse'
require 'yaml'

options = {
  manifest: File.expand_path('../frontend/manifest.json', __dir__),
  release: File.expand_path('releases/android.yaml', __dir__),
  platform: 'android'
}

OptionParser.new do |parser|
  parser.banner = 'Usage: deploy/read-release.rb [options]'
  parser.on('--manifest PATH', 'Path to frontend/manifest.json') { |path| options[:manifest] = path }
  parser.on('--release PATH', 'Path to release YAML') { |path| options[:release] = path }
  parser.on('--platform PLATFORM', 'Release platform') { |platform| options[:platform] = platform }
end.parse!

def manifest_value(manifest, name)
  match = manifest.match(/"#{Regexp.escape(name)}"\s*:\s*"([^"]+)"/)
  abort "Missing #{name} in manifest" unless match && !match[1].strip.empty?

  match[1].strip
end

manifest = File.read(options[:manifest])
version_name = manifest_value(manifest, 'versionName')
version_code = manifest_value(manifest, 'versionCode')
unless version_code.match?(/\A\d+\z/) && version_code.to_i.positive?
  abort "Invalid versionCode: #{version_code.inspect}"
end

release = YAML.safe_load(File.read(options[:release]), permitted_classes: [], aliases: false) || {}
release_notes = release.fetch('release_notes', [])
unless release_notes.is_a?(Array) && release_notes.all? { |note| note.is_a?(String) }
  abort 'release_notes must be a YAML list of strings'
end

puts JSON.generate(
  platform: options[:platform],
  version_code: version_code.to_i,
  version_name: version_name,
  package_file: "leave-#{options[:platform]}-#{version_code}-#{version_name}.apk",
  release_notes: release_notes
)
