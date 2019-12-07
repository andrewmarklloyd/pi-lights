function scheduleSubmit() {
  var onTimeValue = document.getElementById("onTime").value;
  var offTimeValue = document.getElementById("offTime").value;
  if (onTimeValue === "" || offTimeValue === "") {
    alert("Both 'On Time' and 'Off Time' must be filled in.")
    return false
  }
  return true
}

function clearSubmit() {
  document.getElementById("onTime").value = ""
  document.getElementById("offTime").value = ""
}
