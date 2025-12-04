const API_BASE_URL = '/api';
const PAGE_SIZE = 10;

let state = {
    currentPage: 1,
    totalPages: 1,
    totalComments: 0,
    searchQuery: '',
    currentSort: 'created_at',
    currentSortOrder: 'desc',
    selectedParentId: null,
    commentToDelete: null
};

const elements = {
    commentsTree: document.getElementById('commentsTree'),
    searchInput: document.getElementById('searchInput'),
    searchBtn: document.getElementById('searchBtn'),
    clearSearchBtn: document.getElementById('clearSearchBtn'),
    authorInput: document.getElementById('authorInput'),
    contentInput: document.getElementById('contentInput'),
    parentIdInput: document.getElementById('parentIdInput'),
    submitCommentBtn: document.getElementById('submitCommentBtn'),
    prevPageBtn: document.getElementById('prevPageBtn'),
    nextPageBtn: document.getElementById('nextPageBtn'),
    pageInfo: document.getElementById('pageInfo'),
    paginationInfo: document.getElementById('paginationInfo'),
    deleteModal: document.getElementById('deleteModal'),
    confirmDeleteBtn: document.getElementById('confirmDeleteBtn'),
    cancelDeleteBtn: document.getElementById('cancelDeleteBtn'),
    authorCounter: document.getElementById('authorCounter'),
    contentCounter: document.getElementById('contentCounter')
};

function init() {
    loadComments();
    setupEventListeners();
    setupCharCounters();
}

function setupCharCounters() {
    elements.authorInput.addEventListener('input', (e) => {
        elements.authorCounter.textContent = `${e.target.value.length}/50`;
    });
    
    elements.contentInput.addEventListener('input', (e) => {
        elements.contentCounter.textContent = `${e.target.value.length}/1000`;
    });
}

function setupEventListeners() {
    elements.searchBtn.addEventListener('click', () => {
        state.searchQuery = elements.searchInput.value.trim();
        state.currentPage = 1;
        loadComments();
    });
    
    elements.searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            state.searchQuery = elements.searchInput.value.trim();
            state.currentPage = 1;
            loadComments();
        }
    });
    
    elements.clearSearchBtn.addEventListener('click', () => {
        elements.searchInput.value = '';
        state.searchQuery = '';
        state.currentPage = 1;
        loadComments();
    });
    
    elements.submitCommentBtn.addEventListener('click', submitComment);
    
    elements.prevPageBtn.addEventListener('click', () => {
        if (state.currentPage > 1) {
            state.currentPage--;
            loadComments();
        }
    });
    
    elements.nextPageBtn.addEventListener('click', () => {
        if (state.currentPage < state.totalPages) {
            state.currentPage++;
            loadComments();
        }
    });
    
    elements.confirmDeleteBtn.addEventListener('click', confirmDelete);
    elements.cancelDeleteBtn.addEventListener('click', () => {
        elements.deleteModal.style.display = 'none';
        state.commentToDelete = null;
    });
    
    window.addEventListener('click', (e) => {
        if (e.target === elements.deleteModal) {
            elements.deleteModal.style.display = 'none';
            state.commentToDelete = null;
        }
    });
}

async function loadComments() {
    showLoading();
    
    const params = new URLSearchParams({
        page: state.currentPage,
        page_size: PAGE_SIZE,
        sort_by: state.currentSort,
        sort_order: state.currentSortOrder
    });
    
    if (state.searchQuery) {
        params.append('search', state.searchQuery);
    }
    
    if (state.selectedParentId) {
        params.append('parent', state.selectedParentId);
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/comments?${params}`);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        
        state.totalComments = data.total;
        state.totalPages = Math.ceil(data.total / PAGE_SIZE);
        
        updatePagination();
        renderComments(data.comments);
        
    } catch (error) {
        console.error('Error loading comments:', error);
        showError('Ошибка при загрузке комментариев');
    }
}

function renderComments(comments) {
    if (!comments || comments.length === 0) {
        if (state.searchQuery) {
            elements.commentsTree.innerHTML = `
                <div class="no-comments">
                    <i class="fas fa-search"></i>
                    <p>Комментарии по запросу "${state.searchQuery}" не найдены</p>
                </div>
            `;
        } else {
            elements.commentsTree.innerHTML = `
                <div class="no-comments">
                    <i class="fas fa-comment-slash"></i>
                    <p>Комментариев пока нет. Будьте первым!</p>
                </div>
            `;
        }
        return;
    }
    
    let html = '';
    
    function renderComment(comment, level = 0) {
        const indent = level * 40;
        const date = new Date(comment.created_at).toLocaleString('ru-RU');
        
        html += `
            <div class="comment" style="margin-left: ${indent}px" data-id="${comment.id}">
                <div class="comment-header">
                    <div class="comment-author">
                        <i class="fas fa-user"></i> ${escapeHtml(comment.author)}
                    </div>
                    <div class="comment-meta">
                        <span><i class="far fa-clock"></i> ${date}</span>
                        <span><i class="fas fa-hashtag"></i> ID: ${comment.id}</span>
                        ${comment.parent_id ? `<span><i class="fas fa-reply"></i> Ответ на #${comment.parent_id}</span>` : ''}
                    </div>
                </div>
                <div class="comment-content">
                    ${escapeHtml(comment.content)}
                </div>
                <div class="comment-actions">
                    <button class="comment-reply" onclick="replyToComment(${comment.id}, '${escapeHtml(comment.author)}')">
                        <i class="fas fa-reply"></i> Ответить
                    </button>
                    <button class="comment-delete" onclick="showDeleteModal(${comment.id})">
                        <i class="fas fa-trash"></i> Удалить
                    </button>
                </div>
        `;
        
        if (comment.children && comment.children.length > 0) {
            html += '<div class="comment-children">';
            comment.children.forEach(child => renderComment(child, level + 1));
            html += '</div>';
        }
        
        html += '</div>';
    }
    
    comments.forEach(comment => renderComment(comment));
    elements.commentsTree.innerHTML = html;
}

async function submitComment() {
    const author = elements.authorInput.value.trim();
    const content = elements.contentInput.value.trim();
    const parentId = elements.parentIdInput.value.trim();
    
    if (!author || !content) {
        showError('Пожалуйста, заполните все обязательные поля');
        return;
    }
    
    if (author.length > 50) {
        showError('Имя не должно превышать 50 символов');
        return;
    }
    
    if (content.length > 1000) {
        showError('Комментарий не должен превышать 1000 символов');
        return;
    }
    
    const commentData = {
        author: author,
        content: content
    };
    
    if (parentId) {
        const parentIdNum = parseInt(parentId);
        if (!isNaN(parentIdNum)) {
            commentData.parent_id = parentIdNum;
        }
    }
    
    try {
        elements.submitCommentBtn.disabled = true;
        elements.submitCommentBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Отправка...';
        
        const response = await fetch(`${API_BASE_URL}/comments`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(commentData)
        });
        
        if (!response.ok) {
            let errorMessage = 'Ошибка при отправке комментария';
            try {
                const errorData = await response.text();
                errorMessage = errorData || errorMessage;
            } catch (e) {

            }
            throw new Error(errorMessage);
        }
        
        elements.authorInput.value = '';
        elements.contentInput.value = '';
        elements.parentIdInput.value = '';
        elements.authorCounter.textContent = '0/50';
        elements.contentCounter.textContent = '0/1000';
        
        state.selectedParentId = null;
        
        state.currentPage = 1;
        await loadComments();
        
        showSuccess('Комментарий успешно добавлен!');
        
    } catch (error) {
        console.error('Error submitting comment:', error);
        showError(`Ошибка при отправке комментария: ${error.message}`);
    } finally {
        elements.submitCommentBtn.disabled = false;
        elements.submitCommentBtn.innerHTML = '<i class="fas fa-paper-plane"></i> Отправить комментарий';
    }
}

function replyToComment(commentId, authorName) {
    elements.parentIdInput.value = commentId;
    elements.contentInput.focus();
    elements.contentInput.placeholder = `Ответ ${authorName}...`;
    
    elements.contentInput.scrollIntoView({ behavior: 'smooth' });
    
    showSuccess(`Вы отвечаете на комментарий #${commentId}`);
}

function showDeleteModal(commentId) {
    state.commentToDelete = commentId;
    elements.deleteModal.style.display = 'flex';
}

async function confirmDelete() {
    if (!state.commentToDelete) return;
    
    try {
        elements.confirmDeleteBtn.disabled = true;
        elements.confirmDeleteBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Удаление...';
        
        const response = await fetch(`${API_BASE_URL}/comments/${state.commentToDelete}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        elements.deleteModal.style.display = 'none';
        
        await loadComments();
        
        showSuccess('Комментарий успешно удален!');
        
    } catch (error) {
        console.error('Error deleting comment:', error);
        showError('Ошибка при удалении комментария');
    } finally {
        elements.confirmDeleteBtn.disabled = false;
        elements.confirmDeleteBtn.innerHTML = 'Удалить';
        state.commentToDelete = null;
    }
}

function updatePagination() {
    elements.prevPageBtn.disabled = state.currentPage <= 1;
    elements.nextPageBtn.disabled = state.currentPage >= state.totalPages;
    
    elements.pageInfo.textContent = `Страница ${state.currentPage} из ${state.totalPages}`;
    elements.paginationInfo.textContent = `Всего комментариев: ${state.totalComments}`;
    
    if (state.currentPage <= 1) {
        elements.prevPageBtn.classList.add('disabled');
    } else {
        elements.prevPageBtn.classList.remove('disabled');
    }
    
    if (state.currentPage >= state.totalPages) {
        elements.nextPageBtn.classList.add('disabled');
    } else {
        elements.nextPageBtn.classList.remove('disabled');
    }
}

function showLoading() {
    elements.commentsTree.innerHTML = `
        <div class="loading">
            <i class="fas fa-spinner fa-spin"></i> Загрузка комментариев...
        </div>
    `;
}

function showError(message) {
    elements.commentsTree.innerHTML = `
        <div class="error">
            <i class="fas fa-exclamation-circle"></i>
            <p>${message}</p>
            <button class="btn btn-secondary" onclick="loadComments()">
                <i class="fas fa-redo"></i> Попробовать снова
            </button>
        </div>
    `;
}

function showSuccess(message) {
    document.querySelectorAll('.toast').forEach(toast => toast.remove());
    
    const toast = document.createElement('div');
    toast.className = 'toast success';
    toast.innerHTML = `
        <i class="fas fa-check-circle"></i>
        <span>${escapeHtml(message)}</span>
    `;
    
    document.body.appendChild(toast);
    
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: #2ecc71;
        color: white;
        padding: 15px 25px;
        border-radius: 8px;
        box-shadow: 0 4px 15px rgba(46, 204, 113, 0.3);
        display: flex;
        align-items: center;
        gap: 10px;
        z-index: 10000;
        animation: slideIn 0.3s ease-out;
    `;
    
    if (!document.querySelector('#toast-styles')) {
        const style = document.createElement('style');
        style.id = 'toast-styles';
        style.textContent = `
            @keyframes slideIn {
                from { transform: translateX(100%); opacity: 0; }
                to { transform: translateX(0); opacity: 1; }
            }
            @keyframes slideOut {
                from { transform: translateX(0); opacity: 1; }
                to { transform: translateX(100%); opacity: 0; }
            }
        `;
        document.head.appendChild(style);
    }
    
    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease-out forwards';
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300);
    }, 3000);
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

document.addEventListener('DOMContentLoaded', init);