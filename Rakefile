require 'fileutils'
require 'yaml'

sdk_url = 'https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_darwin_amd64-1.9.22.zip'
sdk_file = "dep/#{File.basename sdk_url}"

def get_version()
	v = `git describe`
	g = v.scan(/v(\d+)-(\d+)/)[0]
	return g[0].to_i + g[1].to_i
end

file 'dep' do |t|
    puts 'execute dep'
    mkdir_p t.name
end

file sdk_file => 'dep' do |t|
    sh 'curl', '-o', t.name, sdk_url
end

task :app do |t|
	app = YAML.load(IO.read('cfg/app.yaml'))
	app['version'] = get_version
	File.open('mitml/app.yaml', 'w') { |w|
		w.write(app.to_yaml)
	}
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

task :serve => ['dep/go_appengine', :app] do
	sh 'dep/go_appengine/goapp', 'serve', 'mitml'
end

task :deploy => ['dep/go_appengine', :app] do
	sh 'dep/go_appengine/goapp', 'deploy', 'mitml'
end

task :version do
	puts get_version
end