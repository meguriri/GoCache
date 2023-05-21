$(document).ready(function () {
  if(checkCookie()==1){
    $('#connectLabel').css("background-color","#19a262")
    $('#connectLabel').html('<b>connected</b>')
    //$('#disconnected').css("display","none") 
    $.ajax({
      type: 'post',
      url: '/connect',
      success: function (res) {
        console.log(res)
        $('#connected').css("display","inline")
        //$('#disconnected').css("display","none")
        $('#connectLabel').css("background-color","#19a262")
        $('#connectLabel').html("<b>connected<b>")
      },
      error: function(jqXHR, textStatus, errorThrown){ 
        $('#connected').css("display","none")
      } 
    })
  }else{
    $('#disconnected').css("display","inline")
  }


  $('#connectBtn').click(function(){
    let ip =$('#serverIpInput').val()
    let port =$('#serverPortInput').val()
    if(ip==""){
      alert('Please enter IP'); 
      $('#serverIpInput').val('')
      $('#serverPortInput').val('')
    }else if(port==""){
      alert('Please enter Port'); 
      $('#serverIpInput').val('')
      $('#serverPortInput').val('')
    }else{
      let data ={
        "address":ip+":"+port
      }
      console.log(data)
      $.ajax({
        type: 'post',
        url: '/connect',
        data: data,
        success: function (res) {
          $('#connected').css("display","inline")
          $('#disconnected').css("display","none")
          $('#connectLabel').css("background-color","#19a262")
          $('#connectLabel').html("<b>connected<b>")
        },
        error: function(jqXHR, textStatus, errorThrown){ 
          alert('connect error'); 
          $('#serverIpInput').val('')
          $('#serverPortInput').val('')
        } 
      })
    }
  })

  $('#disconnectBtn').click(function(){
    $.ajax({
      type: 'post',
      url: '/disconnect',
      success: function (res) {
        console.log(res.res)
        $('#serverIpInput').val("")
        $('#serverPortInput').val("")
        $('#connected').css("display","none")
        $('#disconnected').css("display","inline")
        $('#connectLabel').css("background-color","#d53d3d")
        $('#connectLabel').html("<b>disconnected<b>")
        window.location.reload();
      }
    })
  })
})