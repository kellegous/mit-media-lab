require 'fileutils'

sdk_url = 'https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_darwin_amd64-1.9.22.zip'
sdk_file = "dep/#{File.basename sdk_url}"

file 'dep' do |t|
    puts 'execute dep'
    mkdir_p t.name
end

file sdk_file => 'dep' do |t|
    sh 'curl', '-o', t.name, sdk_url
end

file 'dep/go_appengine' => sdk_file do |t|
    sh 'unzip', sdk_file, '-d', 'dep'
    sh 'touch', t.name
end

task :default => 'dep/go_appengine'