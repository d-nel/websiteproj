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
      cache: false,
      contentType: false,
      processData: false,

      xhr: function() {  // Custom XMLHttpRequest
        var myXhr = $.ajaxSettings.xhr();
        if(myXhr.upload){ // Check if upload property exists
          myXhr.upload.addEventListener('progress',progressHandlingFunction, false); // For handling the progress of the upload
        }
        return myXhr;
      },

      success: function (returndata) {
        // set #pid to id in the action="/finalisepost" form
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

function progressHandlingFunction(e) {
  if(e.lengthComputable){
    var inc = e.total / 100;
    var pc = parseInt(e.loaded / inc);

    $('#percent').html(pc + "%");
  }
}
