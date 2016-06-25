var rt

$( document ).ready(function() {
  $("#file").change(function(event) {

    if (this.files && this.files[0]) {
      var reader = new FileReader();

      reader.onload = function (e) {
        $(".loadlabel").css("display", "block");
        $(".post_preview").css("opacity", "0.6");
        $('.post_preview').attr('src', e.target.result);
      }

      reader.readAsDataURL(this.files[0]);
    }

    //disable the default form submission
    event.preventDefault();

    var formData = new FormData($("form#newpost")[0]);

    $.ajax({
      url: '/post/create',
      type: 'POST',
      data: formData,
      cache: false,
      contentType: false,
      processData: false,
      success: function (returndata) {
        // set #pid to id in the action="/post/finalise" form
        $("#pid").val(returndata.PID)
        $(".loadlabel").css("display", "none");
        $(".post_preview").css("opacity", "1.0");
        //$(".post_preview").attr('src', "/posts/"+returndata.PID+"_1024.jpeg")
      }
    });

    $("#filelabel").hide();

    return false;
  });


  $("#reply").keyup(function(event) {
    $(".reply_img").attr('style', "background: url('/posts/" +
    $("#reply").val() + "_preview.jpeg?t=" + new Date().getTime() +
    "') no-repeat;background-size:cover;")

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
