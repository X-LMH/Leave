#!/usr/bin/env ruby
require_relative 'release'

begin
  abort 'Usage: build-version-sql.rb RELEASE_METADATA.json' unless ARGV.length == 1
  puts Release.version_sql(JSON.parse(File.read(ARGV.fetch(0))))
rescue StandardError => e
  warn e.message
  exit 1
end
