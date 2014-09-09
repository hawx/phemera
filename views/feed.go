package views

const feed = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>{{Title}}</title>
    <link>{{Url}}</link>
    <description>{{SafeDesc}}</description>

    {{#Entries}}
      <item>
        <description><![CDATA[{{{Rendered}}}]]></description>
        <pubDate>{{RssTime}}</pubDate>
        <link>{{Url}}#{{Time}}</link>
      </item>
    {{/Entries}}
  </channel>
</rss>`
