$( document ).ready(function() {
  $("#username_input").keyup(function(event) {
    if (usernameOK($("#username_input").val())) {
      $("#username_error").hide();
      $('input[type="submit"]').removeAttr('disabled');
      $("#username_input").removeClass("input_error");
    } else {
      $("#username_error").html("Username may only contain letters and underscores");
      $("#username_error").show();
      $('input[type="submit"]').attr('disabled','disabled');
      $("#username_input").addClass("input_error");
    }

    return false;
  });

});

function usernameOK(username) {
  var ok = username.search("\\W");

  if (ok == -1) {
    return true;
  } else {
    return false;
  }
}
