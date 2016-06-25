package views

const list = `<ul class="posts">
  {{#Entries}}
    <li id="{{Time}}">
      <a class="link" href="#{{Time}}">&#x2020;</a>
      {{{Rendered}}}
    </li>
  {{/Entries}}
</ul>

<footer>
  {{{Description}}}

  {{#LoggedIn}}
    <a href="/add">Add</a>
    <a href="/sign-out">Sign-out</a>
  {{/LoggedIn}}

  {{^LoggedIn}}
    <a href="/sign-in">Sign-in</a>
  {{/LoggedIn}}
</footer>

<script src="/assets/list.js"></script>`
