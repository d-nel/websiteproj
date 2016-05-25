function openTab(event, tab) {
  $("#tabscontent").children().each(function() {
    $(this).hide();
  });

  $("#tabs").children().each(function() {
    $(this).removeClass("selected");
  });

  $(event.currentTarget).addClass("selected");
  $(tab).show();
}
