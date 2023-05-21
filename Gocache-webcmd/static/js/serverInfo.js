$(document).ready(function () {
  if (checkCookie()==1){
    $('#connectLabel').css("background-color","#19a262")
    $('#connectLabel').html('<b>connected</b>')
    serverInfo()
    $('#refresh').click(function(){
      $.ajax({
        type: 'get',
        url: '/server/refresh',
        success: function (res) {
          $('#usedbytes').html(res.usedBytes+" Bytes / "+res.totalBytes+" Bytes in used")
          $('#usedbytesbar').css('width',(res.usedBytes/res.totalBytes*100)+'%')
          $('#usedpeer').html(res.peersNumber+" peers")
          $('#RAM').html("RAM &nbsp;:"+res.usedBytes+" Bytes")
        }
      })
    })
  }else{
    alert("please connect a server")
    window.location.href="/"
  }
})