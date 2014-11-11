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

<script>
function highlight() {
    var els = document.getElementsByClassName("highlight");
    Array.prototype.forEach.call(els, function(el) {
        el.setAttribute("class", "");
    });

    var hash = window.location.hash.slice(1);
    if (hash != "") {
        var el = document.getElementById(hash);
        el.setAttribute("class", "highlight");
    }
}

function gotAssertion(assertion) {
    // got an assertion, now send it up to the server for verification
    if (assertion !== null) {
        $.ajax({
            type: 'POST',
            url: '/sign-in',
            data: { assertion: assertion },
            success: function(res, status, xhr) {
                window.location.reload();
            },
            error: function(xhr, status, res) {
                alert("sign-in failure" + res);
            }
        });
    }
};

jQuery(function($) {
    $('#browserid').click(function() {
        navigator.id.get(gotAssertion);
    });

    highlight();
    window.onhashchange = highlight;
});
</script>`