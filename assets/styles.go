package assets

const Styles = `@import url(http://fonts.googleapis.com/css?family=Lora);

body {
    max-width: 24em;
    margin: 2em;
    background: #fefefa;
    color: #1d1d1a;

    font: 21px/1.4 Lora, serif;
}

.posts {
    list-style: none;
    padding: 0;
}

.posts > li {
    padding-left: 1em;
    margin: 2.8em 0;
    margin-left: -1em;
}

.posts > li .link {
    position: absolute;
    margin-left: -1em;
    color: #acacaa;
    text-decoration: none;
    opacity: 0;
}

.posts > li:hover .link, .posts > li.highlight .link {
    opacity: .99;
}

form textarea {
    font: 21px/1.4 Lora, serif;
    display: block;
    border: none;
    width: 100%;

    margin: 2.8em 0;
    position: relative;
    background: transparent;
    padding: 0;

    white-space: pre-wrap;
    word-wrap: break-word;
    box-sizing: border-box;

    overflow: hidden;
    resize: none;
}

form input[type=submit] {
    /* fancy? */
}

#preview {
    position: absolute;
    top: 2em;
    left: 28em;
    opacity: .7;
    width: 24em;
}

footer {
    font-size: 16px;
    color: #4c4c4a;
}

footer a { color: #1d1d1a; }

footer > a {
    margin-left: 1em;
}

@media screen and (max-width: 30em) {
    body {
        margin: 1em;
    }

    footer {
        font-size: 14px;
    }
}

@media screen and (max-width: 72em) {
    #preview, #preview-btn { display: none; }
}`
