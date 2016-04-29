$( document ).ready(function() {
  $("#file").change(function(event){

    //disable the default form submission
    event.preventDefault();

    var formData = new FormData($('form#newpfp')[0]);

    $.ajax({
      url: '/newpfp',
      type: 'POST',
      data: formData,
      async: false,
      cache: false,
      contentType: false,
      processData: false,
      success: function (returndata) {

        $('.profile_pic').each(function(){
           $(this).attr('src', $(this).attr('src') + "?timestamp=" + new Date().getTime());
        });

      }
    });



    return false;
  });

  $("#postfile").change(function(event){

    //disable the default form submission
    event.preventDefault();

    var formData = new FormData($("form#newpost")[0]);

    $.ajax({
      url: '/newpost',
      type: 'POST',
      data: formData,
      async: false,
      cache: false,
      contentType: false,
      processData: false,
      success: function (returndata) {

      }
    });



    return false;
  });

});