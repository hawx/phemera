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
    <a id="browserid" href="#" title="Sign-in with Persona">Sign-in</a>
  {{/LoggedIn}}
</footer>

<script src="http://code.jquery.com/jquery-2.1.1.min.js"></script>
<script src="https://login.persona.org/include.js"></script>
<script src="/assets/list.js"></script>`
