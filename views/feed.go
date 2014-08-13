package views

const feed = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>{{Title}}</title>
    <link>{{Url}}</link>
    <description><![CDATA[{{{Description}}}]]></description>

    {{#Entries}}
      <item>
        <description><![CDATA[{{{Rendered}}}]]></description>
        <pubDate>{{RssTime}}</pubDate>
      </item>
    {{/Entries}}
  </channel>
</rss>`
