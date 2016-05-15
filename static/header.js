var show = false

$( document ).ready(function() {
  checkNavDisplay()

  $("#navbutton").on("click", function(event) {
    $("#nav").toggle();
    show = !show
  });

  $(window).resize(function(event) {
    checkNavDisplay()
  });
});

function checkNavDisplay() {
  var w = window.innerWidth
  if (w <= 570) {
    if (show) {
      $("#nav").show();
    } else {
      $("#nav").hide();
    }

  } else {
    $("#nav").show();
    show = false
  }
}
