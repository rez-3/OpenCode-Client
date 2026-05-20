// ============================================================
// 站内文件浏览器 - 目录浏览与状态管理
// ============================================================

window.fileBrowserState = {
    rootDir: '',
    currentPath: '/',
    parentPath: '/',
    items: [],
    selectedItem: null,
    previewMeta: null,
    previewContent: null,
    loadingList: false,
    loadingPreview: false,
};

function openFileBrowserModal(rootDir) {
    var modal = document.getElementById('fileBrowserModal');
    var title = document.getElementById('fileBrowserTitle');
    if (!modal) return;
    window.fileBrowserState.rootDir = rootDir || '';
    window.fileBrowserState.currentPath = '/';
    window.fileBrowserState.parentPath = '/';
    window.fileBrowserState.selectedItem = null;
    if (title) title.textContent = '文件浏览 - ' + (rootDir || '');
    modal.style.display = 'flex';
    loadFileBrowserList('/');
}

function closeFileBrowserModal() {
    var modal = document.getElementById('fileBrowserModal');
    if (modal) modal.style.display = 'none';
    clearFileBrowserPreview();
}

function clearFileBrowserPreview() {
    var titleEl = document.getElementById('filePreviewTitle');
    var metaEl = document.getElementById('filePreviewMeta');
    var bodyEl = document.getElementById('filePreviewBody');
    if (titleEl) titleEl.textContent = '请选择文件';
    if (metaEl) metaEl.textContent = '';
    if (bodyEl) bodyEl.innerHTML = '<div class="file-browser-empty">请选择左侧文件进行预览</div>';
}

async function loadFileBrowserList(path) {
    var state = window.fileBrowserState;
    var listEl = document.getElementById('fileBrowserList');
    var emptyEl = document.getElementById('fileBrowserListEmpty');
    if (!listEl || !state.rootDir) return;
    state.loadingList = true;
    state.currentPath = path || '/';
    if (listEl) listEl.innerHTML = '<div class="file-browser-empty">正在读取目录...</div>';
    if (emptyEl) emptyEl.style.display = 'none';
    clearFileBrowserPreview();
    try {
        var resp = await fetch('/api/files/list?rootDir=' + encodeURIComponent(state.rootDir) + '&path=' + encodeURIComponent(state.currentPath));
        if (!resp.ok) throw new Error('列目录失败');
        var data = await resp.json();
        state.currentPath = data.currentPath || '/';
        state.parentPath = data.parentPath || '/';
        state.items = data.items || [];
        renderFileBrowserBreadcrumb();
        renderFileBrowserList();
    } catch (err) {
        if (listEl) listEl.innerHTML = '<div class="file-browser-empty error">' + escapeHtml(err.message || err) + '</div>';
    } finally {
        state.loadingList = false;
    }
}

function renderFileBrowserBreadcrumb() {
    var pathEl = document.getElementById('fileBrowserPath');
    var upBtn = document.getElementById('btnFileBrowserUp');
    var state = window.fileBrowserState;
    if (pathEl) {
        var parts = state.currentPath.split('/').filter(Boolean);
        var html = '<span class="file-browser-crumb" data-path="/">根目录</span>';
        var acc = '/';
        parts.forEach(function(part) {
            acc += part + '/';
            html += '<span class="file-browser-crumb-sep">/</span><span class="file-browser-crumb" data-path="' + escapeHtml(acc) + '">' + escapeHtml(part) + '</span>';
        });
        pathEl.innerHTML = html;
        pathEl.querySelectorAll('[data-path]').forEach(function(node) {
            node.addEventListener('click', function() {
                loadFileBrowserList(this.dataset.path || '/');
            });
        });
    }
    if (upBtn) upBtn.disabled = (state.currentPath === '/' || state.currentPath === '');
}

function renderFileBrowserList() {
    var listEl = document.getElementById('fileBrowserList');
    var emptyEl = document.getElementById('fileBrowserListEmpty');
    var state = window.fileBrowserState;
    if (!listEl) return;
    if (!state.items.length) {
        listEl.innerHTML = '';
        if (emptyEl) {
            emptyEl.textContent = '当前目录为空';
            emptyEl.style.display = 'block';
        }
        return;
    }
    if (emptyEl) emptyEl.style.display = 'none';
    listEl.innerHTML = state.items.map(function(item) {
        var icon = item.type === 'dir' ? '📁' : '📄';
        return '<button type="button" class="file-browser-item ' + (item.type === 'dir' ? 'dir' : 'file') + '" data-path="' + escapeHtml(item.path) + '" data-type="' + escapeHtml(item.type) + '">' +
            '<span class="file-browser-item-icon">' + icon + '</span>' +
            '<span class="file-browser-item-name">' + escapeHtml(item.name) + '</span>' +
        '</button>';
    }).join('');
    listEl.querySelectorAll('.file-browser-item').forEach(function(btn) {
        btn.addEventListener('click', function() {
            var path = this.dataset.path || '/';
            var type = this.dataset.type || 'file';
            if (type === 'dir') {
                loadFileBrowserList(path);
            } else {
                state.selectedItem = state.items.find(function(item) { return item.path === path; }) || null;
                renderFileBrowserSelection();
                renderFilePreview(state.selectedItem);
            }
        });
    });
}

function renderFileBrowserSelection() {
    var listEl = document.getElementById('fileBrowserList');
    var state = window.fileBrowserState;
    if (!listEl) return;
    listEl.querySelectorAll('.file-browser-item').forEach(function(node) {
        node.classList.toggle('active', !!state.selectedItem && node.dataset.path === state.selectedItem.path);
    });
}

function goFileBrowserUp() {
    var state = window.fileBrowserState;
    if (!state.currentPath || state.currentPath === '/') return;
    loadFileBrowserList(state.parentPath || '/');
}

function refreshFileBrowser() {
    var state = window.fileBrowserState;
    loadFileBrowserList(state.currentPath || '/').then(function() {
        if (state.selectedItem) {
            renderFilePreview(state.selectedItem);
        }
    });
}
