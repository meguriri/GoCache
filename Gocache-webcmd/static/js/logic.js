function checkCookie(){
  var address=getCookie("address");
  if (address!=""){
      return 1;
  }
  else {
      return 0;
  }
}//检测cookie是否存在
function getCookie(cname){
  var name = cname + "=";
  var ca = document.cookie.split(';');
  for(var i=0; i<ca.length; i++) {
      var c = ca[i].trim();
      if (c.indexOf(name)==0) { return c.substring(name.length,c.length); }
  }
  return "";
}

function serverInfo(){
  $.ajax({
    type: 'get',
    url: '/server/info',
    success: function (res) {
      console.log(res)
      $('#ServerIp').html(res.ip)
      $('#ServerPort').html(res.port)
      $('#ServerPeers').html(res.peers)
      $('#ServerPolicy').html(res.policy)
      $('#address').html('<b>'+res.ip+':'+res.port+'</b>')

      $('#usedbytes').html(res.usedBytes+" Bytes / "+res.totalBytes+" Bytes in used")
      $('#usedbytesbar').css('width',(res.usedBytes/res.totalBytes*100)+'%')
      $('#usedpeer').html(res.peers+" peers")
    }
  })
}
function getPeerList(){
  $.ajax({
      type: 'get',
      url: '/peers/list',
      success: function (res) {
        $('#peerlist').empty()
        for(let i=0;i<res.list.length;i++){
          console.log(res.list[i])
          $('#peerlist').append('<tr>\
          <td class="py-3" style="color:rgb(20, 175, 199);"><h5><b>'+res.list[i].name+'</b></h5></td>\
          <td id="address" class="py-3"><h5>'+res.list[i].address+'</h5></td>\
          <td class="py-3"><h5>'+res.list[i].usedBytes+'</h5></td>\
          <td class="py-3"><h5>'+res.list[i].cacheBytes+'</h5></td>\
          <td class="py-3 ">\
            <a id="delete"  type="button" style="color:rgb(99, 99, 99);"><h4><i class="bi bi-trash-fill me-3"></i></h4></a>\
          </td>\
        </tr>')
        }
      }
    })
}

function getCachesList(address){
  $.ajax({
    type: 'get',
    url: '/caches/list?address='+address,
    success: function (res) {
      if(res.code==200){
        console.log("res:",res)
        $('#cacheAddress').html("<b>"+address+" Caches List</b>")
        $('#cacheList').empty()
        $('#addressInput').val('')
        for(let i=0;i<res.list.length;i++){
          $('#cacheList').append('<tr>\
          <td id="key" class="py-3"><h5><b>'+res.list[i].key+'</b></h5></td>\
          <td id="value" class="py-3"><h5><b>'+res.list[i].value+'</b></h5></td>\
          <td class="py-3 ">\
          <a id="update"  type="button" style="color:rgb(99, 99, 99);" \
          data-bs-toggle="modal" data-bs-target="#myModal"><h4><i class="bi bi-arrow-clockwise me-3"></i></h4></i></a>\
          <a id="delete"  type="button" style="color:rgb(99, 99, 99);" \
          ><h4><i class="bi bi-trash-fill me-3"></i></h4></a></td></tr>')
        }
      }else{
        alert(res.error)
      }
    }
  })
}