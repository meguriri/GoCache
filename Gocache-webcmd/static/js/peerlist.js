$(document).ready(function () {
  if (checkCookie()==1){
    $('#connectLabel').css("background-color","#19a262")
    $('#connectLabel').html('<b>connected</b>')
    getPeerList()

    $('#newConnect').click(function(){
      if($('#nameIpInput').val()==""){
        alert('Please enter name')
      }else if($('#addressIpInput').val()==""){
        alert('Please enter address')
      }else if($('#maxBytesPortInput').val()==""){
        alert('Please enter maxBytes')
      }else{
        let data ={
          "name":$('#nameIpInput').val(),
          "address":$('#addressIpInput').val(),
          "maxBytes":$('#maxBytesPortInput').val(),
        }
        $.ajax({
          type: 'post',
          url: '/peers/connect',
          data:data,
          success: function (res) {
            console.log(res)
            $('#myModal').modal('hide');
            getPeerList()
          },
          error: function(jqXHR, textStatus, errorThrown){ 
            alert('connect error'); 
            $('#nameIpInput').val('')
            $('#addressIpInput').val('')
            $('#maxBytesInput').val('')
          } 
        })
      }
    })
    
    $('#refresh').click(function(){
      $.ajax({
        type: 'get',
        url: '/peers/refresh',
        success: function (res) {
          console.log(res)
          $('#usedbytes').html(res.usedBytes+" Bytes / "+res.totalBytes+" Bytes in use")
          $('#usedbytesbar').css('width',(res.usedBytes/res.totalBytes*100)+'%')
          $('#usedpeer').html(res.peersNumber+" peers")
          $('#RAM').html("RAM &nbsp;:"+res.usedBytes+" Bytes")
        }
      })
    })

    $('#peerlist').on('click','#delete',function(){
      let address=$(this).parent().parent().children('#address').text()
      console.log(address)
      $.ajax({
        type: 'post',
        url: '/peers/delete',
        data: {"address":address},
        success: function (res) {
          console.log(res)
          if (res.res=='there is only one peer,kill failed'){
            alert(res.res)
          }else{
            getPeerList()
          }
        }
      })
    })
  }else{
    alert("please connect a server")
    window.location.href="/"
  }
})