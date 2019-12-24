function scheduleSubmit() {
  var onTimeValue = document.getElementById("onTime").value;
  var offTimeValue = document.getElementById("offTime").value;
  if (onTimeValue === "" || offTimeValue === "") {
    alert("Both 'On Time' and 'Off Time' must be filled in.")
    return false
  }
  return true
}

function autoUpdates() {
  setTimeout(() => {
    location.reload()
  }, 500)
}

function clearSubmit() {
  document.getElementById("onTime").value = ""
  document.getElementById("offTime").value = ""
}

function systemUpdate() {
  var answer = confirm('Are you sure you want to update to the newest version?')
  if (answer) {
    setTimeout(() => {
      setInterval(() => {
        $.get('/static/latestVersion')
          .done(function (data) {
            location.reload();
          })
      }, 2000)
    }, 10000)

    $('#app').hide(1000)
    $('#navbar').hide(1000)
    $( "#subtitle" ).html("Updating to latest version. Page will reload when finished.")
    $( "#title" ).css( "border", "3px solid red" );
    var dots = 0;
    var animation = setInterval(() => {
      if(dots < 8) {
          for (var i = 0; i < dots; i++) {
            $('#dots').append('.');
          }
          $('#dots').append('.');
          $('#dots').append('<br>');
          dots++;
      } else {
          $('#dots').html('');
          dots = 0;
      }
    }, 600);
  }
}
