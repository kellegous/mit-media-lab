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

file 'mitml/app.yaml' => ['cfg/app.yaml'] do |t|
end

file 'dep/go_appengine' => sdk_file do |t|
    sh 'unzip', sdk_file, '-d', 'dep'
    sh 'touch', t.name
end

task :test => ['dep/go_appengine'] do
	s = Process.wait Process.spawn({},
		'../dep/go_appengine/goapp',
		'test',
		:chdir=> 'mitml')
	fail "command status: #{s}" unless $?.exitstatus == 0
end

task :serve => ['dep/go_appengine'] do
	sh 'dep/go_appengine/goapp', 'serve', 'mitml'
end

task :deploy => ['dep/go_appengine'] do
	sh 'dep/go_appengine/goapp', 'deploy', 'mitml'
end