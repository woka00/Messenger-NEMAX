package httpapi

const (
	loginHTML = `

<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <title>Вход — Мессенджер</title>
  <style>
    body { margin:0; background:#0f172a; color:#e5e7eb; font-family:system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif; display:flex; align-items:center; justify-content:center; min-height:100vh; }
    .card { background:#020617; border:1px solid #111827; border-radius:12px; padding:24px; width:360px; box-shadow:0 20px 40px rgba(15,23,42,0.8); }
    h1 { margin:0 0 12px; font-size:20px; }
    p { margin:0 0 16px; color:#9ca3af; font-size:13px; }
    label { display:block; font-size:12px; color:#9ca3af; margin-bottom:6px; }
    input { width:100%; padding:10px; border-radius:10px; border:1px solid #374151; background:#020617; color:#e5e7eb; }
    input:focus { outline:none; border-color:#3b82f6; box-shadow:0 0 0 1px rgba(59,130,246,0.4); }
    button { width:100%; margin-top:12px; padding:10px; border:none; border-radius:10px; background:#3b82f6; color:white; font-weight:600; cursor:pointer; }
    button:hover { background:#2563eb; }
    .link { margin-top:10px; text-align:center; font-size:13px; color:#9ca3af; }
    .link a { color:#60a5fa; text-decoration:none; }
    .status { font-size:12px; min-height:16px; margin-top:8px; }
    .ok { color:#34d399; }
    .err { color:#f97373; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Вход</h1>
    <p>Введите логин и пароль. Если нет аккаунта — зарегистрируйтесь.</p>
    <label for="login">Логин</label>
    <input id="login" placeholder="admin1">
    <label for="password">Пароль</label>
    <input id="password" type="password" placeholder="admin1">
    <button onclick="doLogin()">Войти</button>
    <div id="status" class="status"></div>
    <div class="link">Нет аккаунта? <a href="/register">Зарегистрироваться</a></div>
  </div>
  <script>
		async function doLogin() {
			const login = document.getElementById('login').value.trim();
			const password = document.getElementById('password').value.trim();
			const status = document.getElementById('status');
			status.textContent = 'Проверяю...'; status.className = 'status';
			if (!login || !password) { status.textContent = 'Заполните логин и пароль'; status.className = 'status err'; return; }
			try {
				// schitaem sha256+base64 vmesto plain
				const enc = new TextEncoder();
				const buf = await crypto.subtle.digest('SHA-256', enc.encode(password));
				let binary = '';
				const bytes = new Uint8Array(buf);
				for (let i = 0; i < bytes.byteLength; i++) binary += String.fromCharCode(bytes[i]);
				const passB64 = btoa(binary);

				const res = await fetch('/api/login', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({login,password:passB64}), credentials:'include' });
				if (res.ok) {
					status.textContent = 'Успех! Перенаправляю...'; status.className = 'status ok';
					window.location.href = '/app';
				} else {
					let txt = '';
					try {
						const ct = res.headers.get('content-type') || '';
						if (ct.includes('application/json')) {
							const data = await res.json();
							txt = data.error || data.message || '';
						} else {
							txt = await res.text();
						}
					} catch (e) { txt = ''; }
					if (!txt) {
						if (res.status === 429) txt = 'Слишком много попыток. Попробуйте позже.';
						else if (res.status === 401) txt = 'Неверный логин или пароль';
						else txt = 'Ошибка входа. Код: ' + res.status;
					}
					status.textContent = txt; status.className = 'status err';
				}
			} catch (e) { status.textContent = 'Ошибка соединения'; status.className = 'status err'; }
		}
    document.addEventListener('keydown', (e)=>{ if(e.key==='Enter') doLogin(); });
  </script>
</body>
</html>

`

	registerHTML = `

<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <title>Регистрация — Мессенджер</title>
  <style>
    body { margin:0; background:#0f172a; color:#e5e7eb; font-family:system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif; display:flex; align-items:center; justify-content:center; min-height:100vh; }
    .card { background:#020617; border:1px solid #111827; border-radius:12px; padding:24px; width:360px; box-shadow:0 20px 40px rgba(15,23,42,0.8); }
    h1 { margin:0 0 12px; font-size:20px; }
    p { margin:0 0 16px; color:#9ca3af; font-size:13px; }
    label { display:block; font-size:12px; color:#9ca3af; margin-bottom:6px; }
    input { width:100%; padding:10px; border-radius:10px; border:1px solid #374151; background:#020617; color:#e5e7eb; }
    input:focus { outline:none; border-color:#3b82f6; box-shadow:0 0 0 1px rgba(59,130,246,0.4); }
    button { width:100%; margin-top:12px; padding:10px; border:none; border-radius:10px; background:#22c55e; color:#022c16; font-weight:700; cursor:pointer; }
    button:hover { background:#16a34a; color:#052e16; }
    .link { margin-top:10px; text-align:center; font-size:13px; color:#9ca3af; }
    .link a { color:#60a5fa; text-decoration:none; }
    .status { font-size:12px; min-height:16px; margin-top:8px; }
    .ok { color:#34d399; }
    .err { color:#f97373; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Регистрация</h1>
    <p>Создайте новый аккаунт. Минимум 3 символа.</p>
    <label for="login">Логин</label>
    <input id="login" placeholder="user123">
    <label for="password">Пароль</label>
    <input id="password" type="password" placeholder="пароль">
    <label for="confirm">Повторите пароль</label>
    <input id="confirm" type="password" placeholder="пароль ещё раз">
    <button onclick="doRegister()">Создать аккаунт</button>
    <div id="status" class="status"></div>
    <div class="link">Уже есть аккаунт? <a href="/">Войти</a></div>
  </div>
  <script>
    async function doRegister() {
      const login = document.getElementById('login').value.trim();
      const password = document.getElementById('password').value.trim();
      const confirm = document.getElementById('confirm').value.trim();
      const status = document.getElementById('status');
      status.textContent = 'Проверяю...'; status.className = 'status';
      if (!login || !password) { status.textContent = 'Заполните все поля'; status.className='status err'; return; }
      if (login.length < 3 || password.length < 3) { status.textContent = 'Минимум 3 символа'; status.className='status err'; return; }
      if (password !== confirm) { status.textContent = 'Пароли не совпадают'; status.className='status err'; return; }
      try {
				// parol ne shlem vchistuyu sha256->base64
				const enc = new TextEncoder();
				const buf = await crypto.subtle.digest('SHA-256', enc.encode(password));
				let binary = '';
				const bytes = new Uint8Array(buf);
				for (let i = 0; i < bytes.byteLength; i++) binary += String.fromCharCode(bytes[i]);
				const passB64 = btoa(binary);

				const res = await fetch('/api/register', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({login,password:passB64}), credentials:'include' });
        if (res.ok) {
          // posle registr srazu loginimsya
					const loginRes = await fetch('/api/login', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({login,password:passB64}), credentials:'include' });
          if (loginRes.ok) {
            status.textContent = 'Аккаунт создан! Перенаправляю...'; status.className='status ok';
            window.location.href = '/app';
          } else {
            status.textContent = 'Аккаунт создан, но не удалось войти. Войдите вручную.'; status.className='status err';
          }
        } else {
          const txt = await res.text();
          status.textContent = txt || 'Не удалось создать аккаунт'; status.className='status err';
        }
      } catch (_) { status.textContent = 'Ошибка соединения'; status.className='status err'; }
    }
    document.addEventListener('keydown', (e)=>{ if(e.key==='Enter') doRegister(); });
  </script>
</body>
</html>

`

	appHTML = `

<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Мессенджер</title>
    <style>
        body {
            font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
            background: #0f172a;
            color: #e5e7eb;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .app {
            background: #020617;
            border-radius: 16px;
            padding: 16px 18px;
            width: 780px;
            max-width: 96vw;
            height: 520px;
            box-shadow: 0 20px 40px rgba(15,23,42,0.8);
            border: 1px solid #1f2937;
            display: flex;
            flex-direction: column;
        }
        .app-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .app-title {
            font-size: 18px;
            font-weight: 600;
            color: #f9fafb;
        }
        .app-subtitle {
            font-size: 12px;
            color: #9ca3af;
        }
        .layout {
            flex: 1;
            display: grid;
            grid-template-columns: 230px 1fr;
            gap: 10px;
            min-height: 0;
        }
        .sidebar {
            background: #020617;
            border-radius: 12px;
            border: 1px solid #111827;
            display: flex;
            flex-direction: column;
            padding: 10px;
        }
        .section-title {
            font-size: 11px;
            text-transform: uppercase;
            letter-spacing: .08em;
            color: #6b7280;
            margin-bottom: 6px;
        }
        .field { margin-bottom: 8px; }
        label {
            display: block;
            font-size: 12px;
            color: #9ca3af;
            margin-bottom: 4px;
        }
        input, textarea, button {
            font-family: inherit;
        }
        textarea {
            width: 100%;
            padding: 8px 10px;
            border-radius: 8px;
            border: 1px solid #374151;
            background: #020617;
            color: #e5e7eb;
            font-size: 13px;
            outline: none;
            resize: vertical;
            min-height: 38px;
            max-height: 80px;
        }
        textarea:focus {
            border-color: #3b82f6;
            box-shadow: 0 0 0 1px rgba(59,130,246,0.4);
        }
        button {
            border: none;
            border-radius: 999px;
            padding: 8px 16px;
            font-size: 13px;
            font-weight: 500;
            cursor: pointer;
            background: #3b82f6;
            color: white;
            transition: background .15s, transform .05s, box-shadow .15s;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 6px;
        }
        button.secondary {
            background: #111827;
            color: #e5e7eb;
            border: 1px solid #374151;
        }
        button:hover {
            background: #2563eb;
        }
        button.secondary:hover {
            background: #1f2937;
        }
        button:active {
            transform: translateY(1px);
            box-shadow: none;
        }
        .status {
            font-size: 12px;
            margin-top: 6px;
            min-height: 16px;
        }
        .status.ok { color: #34d399; }
        .status.err { color: #f97373; }
        .chat-list {
            margin-top: 6px;
            flex: 1;
            overflow-y: auto;
        }
        .chat-item {
            padding: 8px 9px;
            border-radius: 9px;
            border: 1px solid transparent;
            margin-bottom: 5px;
            cursor: pointer;
            display: flex;
            flex-direction: column;
            gap: 2px;
        }
        .chat-item.active {
            background: linear-gradient(135deg, #1d4ed8, #3b82f6);
            border-color: #60a5fa;
        }
        .chat-item:not(.active):hover {
            background: #020617;
            border-color: #1f2937;
        }
        .chat-title {
            font-size: 13px;
            font-weight: 500;
            color: #e5e7eb;
        }
        .chat-subtitle {
            font-size: 11px;
            color: #9ca3af;
        }
        .chat-main {
            background: #020617;
            border-radius: 12px;
            border: 1px solid #111827;
            display: flex;
            flex-direction: column;
            padding: 10px;
        }
        .chat-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 6px;
        }
        .chat-header-title {
            font-size: 14px;
            font-weight: 500;
        }
        .chat-header-info {
            font-size: 11px;
            color: #6b7280;
        }
        .messages {
            margin-top: 4px;
            padding: 8px;
            border-radius: 8px;
            background: #020617;
            border: 1px solid #111827;
            flex: 1;
            overflow-y: auto;
            font-size: 13px;
        }
        .msg-row {
            display: flex;
            margin-bottom: 6px;
        }
        .msg-row.me { justify-content: flex-end; }
        .msg-bubble {
            max-width: 70%;
            padding: 6px 9px;
            border-radius: 12px;
            font-size: 13px;
            line-height: 1.35;
        }
        .msg-row.me .msg-bubble {
            background: #2563eb;
            color: #f9fafb;
            border-bottom-right-radius: 2px;
        }
        .msg-row.other .msg-bubble {
            background: #020617;
            color: #e5e7eb;
            border: 1px solid #1f2937;
            border-bottom-left-radius: 2px;
        }
        .msg-from {
            font-size: 11px;
            margin-bottom: 2px;
            opacity: 0.9;
        }
        .msg-row.me .msg-from { text-align: right; }
        .msg-text { white-space: pre-wrap; word-wrap: break-word; }
        .composer {
            display: flex;
            gap: 8px;
            padding-top: 8px;
        }
    </style>
</head>
<body>
<div class="app">
    <div class="app-header">
        <div>
            <div class="app-title">NEMAX</div>
            <div class="app-subtitle">Лучше, чем телеграм, хуже, чем макс</div>
        </div>
        <div style="display:flex;align-items:center;gap:8px;">
            <div class="app-subtitle" id="currentUserLabel">...</div>
            <button class="secondary" onclick="logout()">Сменить аккаунт</button>
        </div>
    </div>

    <div class="layout">
        <div class="sidebar">
            <div class="section-title">Диалоги</div>
            <div id="chatList" class="chat-list"></div>
        </div>

        <div class="chat-main">
            <div class="chat-header">
                <div class="chat-header-title" id="chatTitle">Выберите диалог</div>
                <div class="chat-header-info" id="chatInfo"></div>
            </div>
            <div id="messages" class="messages">
                <div style="font-size:12px;color:#9ca3af;">Сообщения появятся здесь.</div>
            </div>
            <div class="composer">
                <textarea id="text" placeholder="Напишите сообщение и нажмите Enter или кнопку..."></textarea>
                <button onclick="sendMessage()">Отправить</button>
            </div>
            <div id="sendStatus" class="status"></div>
        </div>
    </div>
</div>

<script>
let currentLogin = null;
let currentPeer = null;
let allUsers = [];
let ws = null;

async function loadCreds() {
    // berem login iz sessii API
    try {
        const res = await fetch('/api/me', { credentials:'include' });
        if (!res.ok) {
            window.location.href = '/';
            return;
        }
        const data = await res.json();
        currentLogin = data.login;
        document.getElementById('currentUserLabel').textContent = 'Вы: ' + currentLogin;
        // posle logina konnektim ws
        connectWebSocket();
    } catch (e) {
        window.location.href = '/';
    }
}

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = protocol + '//' + window.location.host + '/ws';
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        console.log('WebSocket подключен');
    };
    
    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        handleWebSocketMessage(msg);
    };
    
    ws.onerror = (error) => {
        console.error('WebSocket ошибка:', error);
    };
    
    ws.onclose = () => {
        console.log('WebSocket отключен, переподключение через 3 секунды...');
        setTimeout(connectWebSocket, 3000);
    };
}

function handleWebSocketMessage(msg) {
    if (msg.type === 'new_message') {
        // soobshenie pro tekushchiy dialog
        if ((msg.to === currentLogin && msg.from === currentPeer) || 
            (msg.from === currentLogin && msg.to === currentPeer)) {
            // dobavlyaem v chat
            addMessageToDialog(msg.from, msg.text, msg.time);
        }
    }
}

function addMessageToDialog(from, text, time) {
    const box = document.getElementById('messages');
    if (!box) return;
    
    const row = document.createElement('div');
    const me = from === currentLogin;
    row.className = 'msg-row ' + (me ? 'me' : 'other');
    
    const bubble = document.createElement('div');
    bubble.className = 'msg-bubble';
    
    const fromEl = document.createElement('div');
    fromEl.className = 'msg-from';
    fromEl.textContent = me ? 'Вы' : from;
    
    const textEl = document.createElement('div');
    textEl.className = 'msg-text';
    textEl.textContent = text;
    
    bubble.appendChild(fromEl);
    bubble.appendChild(textEl);
    row.appendChild(bubble);
    box.appendChild(row);
    box.scrollTop = box.scrollHeight;
}

async function logout() {
    // zakryvaem ws
    if (ws) {
        ws.close();
        ws = null;
    }
    try {
        await fetch('/api/logout', { method:'POST', credentials:'include' });
    } catch (e) {
        // oshibki ignor
    }
    window.location.href = '/';
}

async function fetchUsers() {
    try {
        const res = await fetch('/api/users', {
            method:'POST',
            credentials:'include'
        });
        if (!res.ok) throw new Error();
        const data = await res.json();
        allUsers = data.users || [];
        buildChatList();
    } catch (_) {
        document.getElementById('chatList').innerHTML = '<div style="font-size:12px;color:#f97373;">Не удалось загрузить пользователей.</div>';
    }
}

function buildChatList() {
    const list = document.getElementById('chatList');
    list.innerHTML = '';
    const peers = allUsers.filter(u => u !== currentLogin);
    if (!peers.length) {
        list.innerHTML = '<div style="font-size:12px;color:#9ca3af;">Диалогов пока нет.</div>';
        return;
    }
    peers.forEach(p => {
        const item = document.createElement('div');
        item.className = 'chat-item' + (p === currentPeer ? ' active' : '');
        item.onclick = () => { currentPeer = p; buildChatList(); loadDialog(); };
        const title = document.createElement('div');
        title.className = 'chat-title';
        title.textContent = 'Диалог с ' + p;
        const subtitle = document.createElement('div');
        subtitle.className = 'chat-subtitle';
        subtitle.textContent = 'Личные сообщения';
        item.appendChild(title);
        item.appendChild(subtitle);
        list.appendChild(item);
    });
    if (!currentPeer && peers.length) currentPeer = peers[0];
}

async function sendMessage() {
    const fromLogin = currentLogin;
    const toLogin = currentPeer;
    const text = document.getElementById('text').value.trim();
    const statusEl = document.getElementById('sendStatus');
    statusEl.textContent = '';
    statusEl.className = 'status';

    if (!fromLogin || !toLogin) {
        statusEl.textContent = 'Выберите собеседника.';
        statusEl.className = 'status err';
        return;
    }
    if (!text) {
        statusEl.textContent = 'Введите текст сообщения.';
        statusEl.className = 'status err';
        return;
    }

    try {
        const res = await fetch('/api/send', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            credentials: 'include',
            body: JSON.stringify({fromLogin, toLogin, text})
        });
        if (res.ok) {
            statusEl.textContent = 'Сообщение отправлено.';
            statusEl.className = 'status ok';
            document.getElementById('text').value = '';
            // ws sam dobavit ne zovem loadDialog
        } else if (res.status === 401) {
            statusEl.textContent = 'Неверный логин или пароль.';
            statusEl.className = 'status err';
            logout();
        } else {
            const t = await res.text();
            statusEl.textContent = 'Ошибка: ' + t;
            statusEl.className = 'status err';
        }
    } catch (e) {
        statusEl.textContent = 'Ошибка соединения с сервером.';
        statusEl.className = 'status err';
    }
}

async function loadDialog() {
    const login = currentLogin;
    const withUser = currentPeer;
    const box = document.getElementById('messages');

    if (!login || !withUser) {
        box.innerHTML = '<div style="font-size:12px;color:#9ca3af;">Выберите диалог.</div>';
        document.getElementById('chatTitle').textContent = 'Выберите диалог';
        document.getElementById('chatInfo').textContent = '';
        return;
    }

    document.getElementById('chatTitle').textContent = 'Диалог с ' + withUser;
    document.getElementById('chatInfo').textContent = 'Сообщения между ' + login + ' и ' + withUser;

    box.innerHTML = '<div style="font-size:12px;color:#9ca3af;">Загружаю...</div>';
    try {
        const res = await fetch('/api/dialog', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            credentials: 'include',
            body: JSON.stringify({login, with: withUser})
        });
        if (!res.ok) {
            box.innerHTML = '<div style="font-size:12px;color:#f97373;">Ошибка / неверные данные.</div>';
            return;
        }
        const data = await res.json();
        if (!data.length) {
            box.innerHTML = '<div style="font-size:12px;color:#9ca3af;">Диалог пока пуст.</div>';
            return;
        }
        box.innerHTML = '';
        data.forEach(m => {
            const row = document.createElement('div');
            const me = m.from === login;
            row.className = 'msg-row ' + (me ? 'me' : 'other');

            const bubble = document.createElement('div');
            bubble.className = 'msg-bubble';

            const from = document.createElement('div');
            from.className = 'msg-from';
            from.textContent = me ? 'Вы' : m.from;

            const text = document.createElement('div');
            text.className = 'msg-text';
            text.textContent = m.text;

            bubble.appendChild(from);
            bubble.appendChild(text);
            row.appendChild(bubble);
            box.appendChild(row);
        });
        box.scrollTop = box.scrollHeight;
    } catch (e) {
        box.innerHTML = '<div style="font-size:12px;color:#f97373;">Ошибка соединения с сервером.</div>';
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    await loadCreds();
    document.getElementById('text').addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });
    fetchUsers().then(loadDialog);
    // avtoobnovlenie ubral dialog gruzim po deystviyam
});
</script>
</body>
</html>

`
)
