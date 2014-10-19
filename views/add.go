package views

const add = `<form action="/add" method="post">
  <textarea autofocus="autofocus" id="body" name="body"></textarea>
  <input type="submit" value="Add" />
  <input type="button" id="preview-btn" value="Preview" />
</form>

<div id="preview"></div>

<script src="http://code.jquery.com/jquery-2.1.1.min.js"></script>
<script src="/assets/jquery.caret.js"></script>
<script src="/assets/jquery.autosize.min.js"></script>

<script>
function scrollToBottom() {
  var pos = $('textarea').caret('pos');
  var len = $('textarea').val().length;

  if (pos === len) {
    window.scrollTo(0, document.body.scrollHeight);
  }
}

function preview() {
  $.post("/preview", $("textarea").val(), function(data, status, xhr) {
    $("#preview").html(data);
  });
}

$(function(){
  $('textarea').autosize({'callback': scrollToBottom});
  $('#preview-btn').click(preview);
});
</script>`
