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

$(function($) {
    highlight();
    connect();

    window.onhashchange = highlight;
});`
