require 'json'
require 'leveldb-native'
require 'nokogiri'
require 'pathname'
require 'redcarpet'
require 'sinatra'
require 'slim'
require 'time'
require 'yaml'
require_relative 'lib/persona'


Settings = YAML.load_file(File.expand_path("../settings.yml", __FILE__))

opts = {
  no_intra_emphasis: true,
  fenced_code_blocks: true,
  disable_indented_code_blocks: true,
  strikethrough: true,
  space_after_headers: true,
  superscript: true,
  underline: true,
  highlight: true
}

smartHtml = Class.new(Redcarpet::Render::HTML) {
  include Redcarpet::Render::SmartyPants
}

Markdown = Redcarpet::Markdown.new(smartHtml, opts)

enable :sessions
set :session_secret, 'HGGejnrgjnakrug4873731yhkjgnr'

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
    authorized? && Settings['users'].include?(authorized_email)
  end

  def logged_in!
    session[:authorize_redirect_url] = request.url
    halt 403 unless logged_in?
  end

  def notify(data)
    $connections.each {|out|
      out << "data: #{data}\n\n"
    }
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

get '/logout' do
  logout!
  redirect '/'
end

get '/feed' do
  content_type 'application/rss+xml'

  Nokogiri::XML::Builder.new {|xml|
    xml.rss(version: '2.0') {
      xml.channel {
        xml.title       'ugh.hawx.me'
        xml.link        'http://ugh.hawx.me'
        xml.description 'a forgetful blog'

        posts.each do |time, body|
          xml.item {
            xml.description { xml.cdata markdown(body) }
            xml.pubDate Time.at(time.to_i).strftime("%a, %d %b %Y %H:%M:%S %z")
          }
        end
      }
    }
  }.to_xml
end

error 400..510 do
  content_type 'text/plain'
  'oops...'
end
