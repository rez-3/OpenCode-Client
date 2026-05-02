// ============================================================
// OpenCode 管理中心 - 常用命令视图
// ============================================================
let commandsData = [];
let commandsLoaded = false;
let cmdActiveTab = 'cli';

async function loadCommands() {
    const content = document.getElementById('cmdContent');
    if (commandsLoaded) return;

    content.innerHTML = '<div class="loading"><div class="spinner"></div><p>正在加载命令列表...</p></div>';

    try {
        const data = await api.GetCommands();
        commandsData = data || [];
        commandsLoaded = true;
        renderCommands(cmdActiveTab);
    } catch (err) {
        content.innerHTML = `<div class="error">
            <p>⚠️ 加载命令失败</p>
            <p class="error-detail">${escapeHtml(err.message || err)}</p>
            <button class="btn btn-primary" onclick="loadCommands()">重试</button>
        </div>`;
    }
}

function renderCommands(tab) {
    const content = document.getElementById('cmdContent');
    const isCLI = tab === 'cli';
    const filtered = commandsData.filter(g => g.isTui === !isCLI);

    if (filtered.length === 0) {
        content.innerHTML = '<div class="empty"><p>📭 没有找到命令</p></div>';
        return;
    }

    content.innerHTML = '';
    filtered.forEach(group => {
        const groupDiv = document.createElement('div');
        groupDiv.className = 'cmd-group';

        const titleEl = document.createElement('h3');
        titleEl.className = 'cmd-group-title';
        titleEl.textContent = group.title;
        groupDiv.appendChild(titleEl);

        const grid = document.createElement('div');
        grid.className = 'cmd-grid';

        (group.cmds || []).forEach(cmd => {
            const card = document.createElement('div');
            card.className = 'cmd-card';

            const top = document.createElement('div');
            top.className = 'cmd-card-top';

            const nameEl = document.createElement('span');
            nameEl.className = 'cmd-card-name';
            nameEl.textContent = cmd.name;

            top.appendChild(nameEl);

            if (cmd.sub) {
                const subEl = document.createElement('span');
                subEl.className = 'cmd-card-sub';
                subEl.textContent = cmd.sub;
                top.appendChild(subEl);
            }

            if (cmd.options) {
                const shortcutEl = document.createElement('span');
                shortcutEl.className = 'cmd-card-shortcut';
                shortcutEl.textContent = cmd.options;
                top.appendChild(shortcutEl);
            }

            card.appendChild(top);

            if (cmd.desc) {
                const descEl = document.createElement('p');
                descEl.className = 'cmd-card-desc';
                descEl.textContent = cmd.desc;
                card.appendChild(descEl);
            }

            grid.appendChild(card);
        });

        groupDiv.appendChild(grid);
        content.appendChild(groupDiv);
    });
}
