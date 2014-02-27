require 'json'
require 'leveldb-native'
require 'nokogiri'
require 'pathname'
require 'redcarpet'
require 'rest_client'
require 'sinatra'
require 'slim'
require 'time'
require 'yaml'

Settings = YAML.load_file(File.expand_path("../settings.yml", __FILE__))

opts = {
  no_intra_emphasis: true,
  fenced_code_blocks: true,
  disable_indented_code_blocks: true,
  strikethrough: true,
  space_after_headers: true,
  superscript: true,
  highlight: true
}

smartHtml = Class.new(Redcarpet::Render::HTML) {
  include Redcarpet::Render::SmartyPants
}

Markdown = Redcarpet::Markdown.new(smartHtml, opts)

enable :sessions
set :session_secret, Settings['secret']

helpers do
  def db(&block)
    path = Pathname.new(File.expand_path("..", __FILE__)) + Settings['db']
    db = LevelDBNative::DB.new(path.to_s)
    yield(db).tap { db.close }
  end

  def save(time, body)
    unixtime = time.to_i.to_s
    db {|db| db.put unixtime, body }

    rendered_body = slim :item, locals: {time: unixtime, body: body}
    notify rendered_body
  end

  def posts
    time = Time.now.to_i - Settings['horizon'].to_i
    db {|db| db.reverse_each(from: time.to_s).entries }
  end

  def logged_in?
    session[:email] && Settings['users'].include?(session[:email])
  end

  def logged_in!
    halt 403 unless logged_in?
  end

  def csrf_token
    session[:csrf]
  end

  def notify(data)
    $connections.each {|out| out << "data: #{data}\n\n" }
  end
end

$connections = []

get '/' do
  slim :list
end

get '/connect', provides: 'text/event-stream' do
  stream(:keep_open) {|out|
    $connections << out
    out.callback { $connections.delete(out) }
  }
end

before('/add') { logged_in! }

get '/add' do
  slim :add
end

post '/add' do
  save Time.now, params['body']
  redirect '/'
end

get '/feed' do
  content_type 'application/rss+xml'
  nokogiri :feed
end

post '/login' do
  # check assertion with a request to the verifier
  response = nil
  if params[:assertion]
    restclient_url = "https://verifier.login.persona.org/verify"
    restclient_params = {
      :assertion => params["assertion"],
      :audience => "http://#{Settings['host']}:#{request.port}",
    }
    response = JSON.parse(RestClient::Resource.new(restclient_url, :verify_ssl => true).post(restclient_params))
  end

  # create a session if assertion is valid
  if response["status"] == "okay"
    session[:email] = response["email"]
    response.to_json
  else
    {:status => "error"}.to_json
  end
end

get '/logout' do
  session[:email] = nil
  redirect '/'
end

error 400..510 do
  content_type 'text/plain'
  'oops...'
end
