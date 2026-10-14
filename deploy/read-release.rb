#!/usr/bin/env ruby
require 'optparse'
require_relative 'release'

options = { root: Release::ROOT, format: 'json' }
OptionParser.new do |p|
  p.on('--tag TAG') { |v| options[:tag] = v }
  p.on('--root PATH') { |v| options[:root] = File.expand_path(v) }
  p.on('--format FORMAT', %w[json github]) { |v| options[:format] = v }
  p.on('--bundle PATH') { |v| options[:bundle] = File.expand_path(v) }
end.parse!
begin
  info = Release.read(options.fetch(:tag), options[:root])
  if options[:bundle]
    Release.bundle(info, options[:bundle], options[:root])
  elsif options[:format] == 'github'
    puts "mode=#{info.fetch('mode')}"
    environment = Release.environment(info)
    puts "environment=#{environment}"
    puts "app_root=#{Release::ENVIRONMENTS.fetch(environment).fetch('root')}"
    puts "admin_dist=#{Release::ENVIRONMENTS.fetch(environment).fetch('admin_dist')}"
    puts "backend_version=#{info.fetch('backend_version')}"
    puts "admin_version=#{info.fetch('admin_version', '')}"
    if info['app']
      puts "apk_path=#{info['app']['apk_path']}"
      puts "apk_name=#{info['app']['apk_name']}"
      puts "version_name=#{info['app']['version_name']}"
      puts "version_code=#{info['app']['version_code']}"
    end
  else
    puts JSON.pretty_generate(info)
  end
rescue StandardError => e
  warn e.message
  exit 1
end
