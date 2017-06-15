function toggleAdminPanel() {
  document.querySelector('#menu').classList.toggle('hidden');
}


document.onkeydown = function(evt) {
  evt = evt || window.event;
  var isEscape = false;
  if("key" in evt) {
    isEscape = (evt.key == "Escape" || evt.key == "Esc");
  } else {
    isEscape = (evt.keyCode == 27);
  }
  if(isEscape) {
    toggleAdminPanel();
  }
}

function showModal(options) {
  var modal = document.getElementById('modal-overlay');
  document.getElementById('modal-title').innerText = (options.title)?options.title:"";
  document.getElementById('modal-subtitle').innerText = (options.subtitle)?options.subtitle:"";
  if(options.body) {
    document.getElementById('modal-body').innerText = options.body;
  } else if(options.bodyNode) {
    document.getElementById('modal-body').appendChild(options.bodyNode);
  }
  if(options.buttons) {
    for(var i = 0; i < options.buttons.length; i++) {
      var btn;
      if(options.buttons[i].isSubmit) {
        btn = document.createElement('submit');
      } else {
        btn = document.createElement('a');
      }
      options.buttons[i].title = (options.buttons[i].title==undefined)?'':options.buttons[i].title;
      options.buttons[i].href = (options.buttons[i].href==undefined)?'#':options.buttons[i].href;
      options.buttons[i].click = (options.buttons[i].click==undefined)?function(){}:options.buttons[i].click;
      options.buttons[i].class = (options.buttons[i].class==undefined)?'':options.buttons[i].class;
      options.buttons[i].position = (options.buttons[i].position==undefined)?'right':options.buttons[i].position;

      btn.innerHTML = options.buttons[i].title;
      btn.title = options.buttons[i].title;
      btn.href = options.buttons[i].href;
      btn.className = 'space pure-button '+options.buttons[i].class+' '+options.buttons[i].position;
      snack.listener(
        {node:btn, event:'click'},
        options.buttons[i].click
      );
      document.getElementById('modal-buttons').appendChild(btn);
    }
  }
  modal.style.visibility = 'visible';
}

function hideModal() {
  var modal = document.getElementById('modal-overlay');
  modal.style.visibility = 'hidden';
  document.getElementById('modal-title').innerHTML = '';
  document.getElementById('modal-body').innerHTML = '';
  var buttonsDiv = document.getElementById('modal-buttons')
  while(buttonsDiv.firstChild) {
    buttonsDiv.removeChild(buttonsDiv.firstChild);
  }
}
