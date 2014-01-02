# 'phemera

A forgetful blog. Posts are markdown. They don't have titles, tags. It has no
features, because when you can't see things that were written a few days ago why
would you need them?

``` bash
$ git clone https://github.com/hawx/phemera.git
$ cd phemera
$ mv settings.yml{-dist,}
```

Then edit settings.yml so users contains your email address(es), and run.

``` bash
$ bundle install
$ bundle exec thin -R config.ru start
...
```
