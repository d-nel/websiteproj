var rt

$( document ).ready(function() {
  $("#newpost").change(function(event){

    //disable the default form submission
    event.preventDefault();

    var formData = new FormData($("form#newpost")[0]);

    $.ajax({
      url: '/createpost',
      type: 'POST',
      data: formData,
      async: false,
      cache: false,
      contentType: false,
      processData: false,
      success: function (returndata) {
        // set #postid to id in the action="/finalisepost" from
        $("#pid").val(returndata.PID)
        $(".post_preview").attr('src', "/posts/"+returndata.PID+"_1024.jpeg")
      }
    });



    return false;
  });


  $("#reply").keyup(function(event) {
    $(".reply_img").attr('style', "background: url('/posts/" +
    $("#reply").val() + "_preview.jpeg?t=" + new Date().getTime() +
    "') no-repeat;")

    hideReplyIfEmpty()

    return false;
  });

  hideReplyIfEmpty()

});

function hideReplyIfEmpty() {
  if ($("#reply").val() == "") {
      $(".reply_preview").hide()
  } else {
    $(".reply_preview").show()
  }
}
