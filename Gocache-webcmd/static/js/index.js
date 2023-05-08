$(document).ready(function () {

  $('#connected').css("display","none")
  $('#connectBtn').click(function(){
    $('#connected').css("display","inline")
    $('#disconnected').css("display","none")
    $('#connectLabel').css("background-color","#19a262")
    $('#connectLabel').html("<b>connect<b>")
  })
  $('#disconnectBtn').click(function(){
    $('#connected').css("display","none")
    $('#disconnected').css("display","inline")
    $('#connectLabel').css("background-color","#d53d3d")
    $('#connectLabel').html("<b>disconnect<b>")
  })
})