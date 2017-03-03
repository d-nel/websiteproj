$( document ).ready(function() {
  $("#popup_button").on("click", function(event) {
    // $("body").toggleClass("noscroll");
    $("#popup").toggle();
  });

  $("#popup_cancel").on("click", function(event) {
    // $("body").removeClass("noscroll");
    $("#popup").hide();
  });
});
