$( document ).ready(function() {
  $("#reg_username").keyup(function(event) {
    if (usernameOK($("#reg_username").val())) {
      $(".register_error").html("")
      $('input[type="submit"]').removeAttr('disabled');

    } else {
      $(".register_error").html("Username may only contain letters and underscores")
      $('input[type="submit"]').attr('disabled','disabled');
    }

    return false;
  });

});

function usernameOK(username) {
  var ok = username.search("\\W");

  if (ok == -1) {
    return true
  } else {
    return false
  }
}
