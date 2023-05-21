$(document).ready(function () {
  var updateKey=""
  var address
  if (checkCookie()==1){
    $('#connectLabel').css("background-color","#19a262")
    $('#connectLabel').html('<b>connected</b>')
    $('#subBtn').click(function(){
      address=$('#addressInput').val()
      console.log("address:",address)
      getCachesList(address)
    })
    
    $('#cacheList').on('click','#delete',function(){
      let key=$(this).parent().parent().children("#key").text()
      console.log(key)
      $.ajax({
        type: 'post',
        url: '/caches/delete',
        data: {"key":key},
        success: function (res) {
          console.log(res)
          getCachesList(address)
        }
      })
    })

    $('#cacheList').on('click','#update',function(){
      updateKey=$(this).parent().parent().children('#key').text()
      $('#valueInput').val('')
    })

    $('#submit').click(function(){
      console.log("updatekay:",updateKey)
      let value=$('#valueInput').val()
      if(value==""){
        alert('Please enter value')
      }else{
        $.ajax({
          type: 'post',
          url: '/caches/update',
          data: {"key":updateKey,"value":value},
          success: function (res) {
            console.log(res)
            $('#myModal').modal('hide');
            getCachesList(address)
          }
        })
      }
    })

    $('#getSubmit').click(function(){
      let key=$('#getInput').val()
      if(key==""){
        alert('Please enter key')
      }else{
        $.ajax({
          type: 'post',
          url: '/caches/get',
          data: {"key":key},
          success: function (res) {
            if(res.value!="(nil)"){
              console.log(res)
              $('#getInput').val('')
              $('#cacheList').empty()
              $('#cacheList').append('<tr>\
              <td id="key" class="py-3"><h5><b>'+res.key+'</b></h5></td>\
              <td id="value" class="py-3"><h5><b>'+res.value+'</b></h5></td>\
              <td class="py-3 ">\
              <a id="update"  type="button" style="color:rgb(99, 99, 99);" \
              data-bs-toggle="modal" data-bs-target="#myModal"><h4><i class="bi bi-arrow-clockwise me-3"></i></h4></i></a>\
              <a id="delete"  type="button" style="color:rgb(99, 99, 99);" \
              ><h4><i class="bi bi-trash-fill me-3"></i></h4></a></td></tr>')
            }else{
              alert("no cahce,key: "+res.key)
              $('#getInput').val('')
            }
          }
        })
      }
    })

    $('#cacheSubmit').click(function(){
      let key=$('#newKeyInput').val()
      let value=$('#newValueInput').val()
      if(value==""){
        alert('Please enter value')
      }else if(key==""){
        alert('Please enter key')
      }else{
        $.ajax({
          type: 'post',
          url: '/caches/set',
          data: {"key":key,"value":value},
          success: function (res) {
            console.log(res)
            $('#cacheModal').modal('hide');
            alert("set ok")
          }
        })
      }
    })
  }else{
    alert("please connect a server")
    window.location.href="/"
  }


})