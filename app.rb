require 'leveldb'
require 'pathname'
require 'redcarpet'
require 'rest-client'
require 'sinatra'
require 'slim'
require 'time'
require 'yaml'
require_relative 'lib/persona'

enable :sessions
set :session_secret, 'HGGejnrgjnakrug4873731yhkjgnr'

helpers do
  def phemera_settings
    @_phemera_settings ||= YAML.load_file(File.expand_path("../settings.yml", __FILE__))
  end

  def db
    @_phemera_db ||= LevelDB::DB.new (Pathname.new(File.expand_path("..", __FILE__)) + phemera_settings['db']).to_s
  end

  def save(time, body)
    db.put time.to_i.to_s, body
  end

  def posts
    db.each(from: (Time.now.to_i - phemera_settings['horizon'].to_i).to_s)
      .entries
      .sort_by(&:first)
      .reverse
  end

  def markdown
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

    Redcarpet::Markdown.new(smartHtml, opts)
  end

  def logged_in?
    authorized? && phemera_settings['users'].include?(authorized_email)
  end

  def logged_in!
    session[:authorize_redirect_url] = request.url
    redirect login_url unless logged_in?
  end
end

get '/' do
  slim :list
end

get '/add' do
  logged_in!
  slim :add
end

post '/add' do
  logged_in!
  save Time.now, params['body']
  redirect '/'
end

get '/logout' do
  logout!
  redirect '/'
end
