$( document ).ready(function() {
  $("#reg_username").keyup(function(event) {
    if (usernameOK($("#reg_username").val())) {
      $(".register_error").html("")
      $('input[type="submit"]').removeAttr('disabled');
      $("#reg_username").removeClass("input_error");

    } else {
      $(".register_error").html("Username may only contain letters and underscores");
      $('input[type="submit"]').attr('disabled','disabled');
      $("#reg_username").addClass("input_error");
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
