function toggleAdminPanel() {
  var menu = document.querySelector('#menu');
  if(menu.classList.contains('hidden')) {
    document.querySelector('#layout>.content').style.marginLeft='150px';
    menu.classList.remove('hidden');
  } else {
    document.querySelector('#layout>.content').style.marginLeft='0';
    menu.classList.add('hidden');
  }
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
  window.scrollTo(0, 0);
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

function setFlashMessage(msg, cls) {
  var flash = document.querySelector('aside.flash');
  flash.innerText = msg;
  for(var i = 0; i < cls.length; i++) {
    flash.classList.add(cls[i]);
  }
  flash.classList.remove('hidden');
  flash.style.opacity=1;
  handleFlashMessage();
}

function handleFlashMessage() {
  var flash = document.querySelector('aside.flash');
  if(flash.classList.contains('fading')) {
    setTimeout(fadeOutFlashMessage, 1000);
  }
}

function fadeOutFlashMessage() {
  var flash = document.querySelector('aside');
  var opac = flash.style.opacity;
  if(opac == "") { opac = 1; }
  if(opac > 0) {
    setTimeout(function() {
      flash.style.opacity = opac - 0.01;
      fadeOutFlashMessage();
    }, 10);
  }
}

handleFlashMessage();
