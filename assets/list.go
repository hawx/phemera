package assets

const List = `function highlight() {
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

function connect() {
    var es = new EventSource('/connect');

    es.onmessage = function(e) {
        var ul = document.getElementsByClassName("posts")[0];
        ul.innerHTML = e.data + ul.innerHTML;
    }
}

function gotAssertion(assertion) {
    // got an assertion, now send it up to the server for verification
    if (assertion !== null) {
        $.ajax({
            type: 'POST',
            url: '/login',
            data: { assertion: assertion, authenticity_token: window.CSRF_TOKEN },
            success: function(res, status, xhr) {
                window.location.reload();
            },
            error: function(xhr, status, res) {
                alert("login failure" + res);
            }
        });
    }
};

jQuery(function($) {
    $('#browserid').click(function() {
        navigator.id.get(gotAssertion);
    });

    highlight();
    connect();

    window.onhashchange = highlight;
});`
