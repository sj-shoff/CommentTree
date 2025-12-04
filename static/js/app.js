class CommentsApp {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 10;
        this.searchQuery = '';
        this.sortBy = 'created_at';
        this.sortOrder = 'desc';
        this.replyingTo = null;
        this.comments = [];
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadComments();
    }

    bindEvents() {
        // Создание комментария
        document.getElementById('createCommentBtn')?.addEventListener('click', (e) => {
            e.preventDefault();
            this.createComment();
        });

        // Поиск
        document.getElementById('searchBtn')?.addEventListener('click', () => this.handleSearch());
        document.getElementById('searchInput')?.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.handleSearch();
        });
        document.getElementById('clearSearch')?.addEventListener('click', () => this.clearSearch());

        // Сортировка
        document.getElementById('sortBy')?.addEventListener('change', () => this.handleSortChange());
        document.getElementById('sortOrder')?.addEventListener('change', () => this.handleSortChange());

        // Пагинация
        document.getElementById('prevPage')?.addEventListener('click', () => this.prevPage());
        document.getElementById('nextPage')?.addEventListener('click', () => this.nextPage());
    }

    async loadComments() {
        const container = document.getElementById('commentsContainer');
        if (!container) return;
        
        container.innerHTML = '<div class="loading">Загрузка комментариев...</div>';

        try {
            const params = new URLSearchParams({
                page: this.currentPage,
                page_size: this.pageSize,
                search: this.searchQuery,
                sort_by: this.sortBy,
                sort_order: this.sortOrder
            });

            const response = await fetch(`/comments?${params}`);
            if (!response.ok) throw new Error('Ошибка загрузки комментариев');
            
            const data = await response.json();
            this.comments = data.comments || [];
            this.renderComments();
            this.updatePagination(data);
            
        } catch (error) {
            this.showError('Ошибка при загрузке комментариев: ' + error.message);
        }
    }

    renderComments() {
        const container = document.getElementById('commentsContainer');
        if (!container) return;
        
        if (this.comments.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-comment-slash"></i>
                    <h4>Комментариев пока нет</h4>
                    <p>Будьте первым, кто оставит комментарий!</p>
                </div>
            `;
            return;
        }

        container.innerHTML = this.comments.map(comment => this.renderCommentNode(comment, 0)).join('');
    }

    renderCommentNode(comment, level = 0) {
        const date = new Date(comment.created_at).toLocaleString('ru-RU');
        const levelClass = level > 0 ? `comment-level-${Math.min(level, 5)}` : '';
        
        let html = `
            <div class="comment-node ${levelClass}" data-comment-id="${comment.id}">
                <div class="comment-header">
                    <span class="comment-author">
                        <i class="fas fa-user"></i> ${this.escapeHtml(comment.author)}
                    </span>
                    <span class="comment-date">
                        <i class="far fa-clock"></i> ${date}
                    </span>
                </div>
                <div class="comment-content">${this.escapeHtml(comment.content)}</div>
                <div class="comment-actions">
                    <button class="btn btn-outline btn-xs reply-btn" 
                            data-comment-id="${comment.id}" 
                            data-author="${this.escapeHtml(comment.author)}">
                        <i class="fas fa-reply"></i> Ответить
                    </button>
                    <button class="btn btn-danger btn-xs delete-btn" 
                            data-comment-id="${comment.id}">
                        <i class="fas fa-trash"></i> Удалить
                    </button>
                </div>
        `;

        // Добавляем дочерние комментарии
        if (comment.children && comment.children.length > 0) {
            html += `<div class="comment-children">`;
            html += comment.children.map(child => this.renderCommentNode(child, level + 1)).join('');
            html += `</div>`;
        }

        html += `</div>`;
        return html;
    }

    async createComment() {
        const author = document.getElementById('commentAuthor')?.value.trim();
        const content = document.getElementById('commentContent')?.value.trim();
        const parentId = document.getElementById('parentId')?.value;

        if (!author || !content) {
            this.showError('Пожалуйста, заполните все поля');
            return;
        }

        const commentData = {
            author: author,
            content: content,
            parent_id: parentId ? parseInt(parentId) : null
        };

        try {
            const response = await fetch('/comments', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(commentData)
            });

            if (!response.ok) throw new Error('Ошибка при создании комментария');

            this.showSuccess('Комментарий успешно добавлен!');
            
            // Очистка формы
            document.getElementById('commentForm')?.reset();
            this.cancelReply();
            
            // Перезагрузка комментариев
            await this.loadComments();

        } catch (error) {
            this.showError('Ошибка: ' + error.message);
        }
    }

    replyTo(commentId, authorName) {
        this.replyingTo = commentId;
        document.getElementById('parentId').value = commentId;
        document.getElementById('replyInfo').style.display = 'block';
        document.getElementById('replyAuthor').textContent = authorName;
        
        const contentInput = document.getElementById('commentContent');
        contentInput.placeholder = `Ответ ${authorName}...`;
        contentInput.focus();
    }

    cancelReply() {
        this.replyingTo = null;
        document.getElementById('parentId').value = '';
        document.getElementById('replyInfo').style.display = 'none';
        document.getElementById('commentContent').placeholder = 'Ваш комментарий...';
    }

    async deleteComment(commentId) {
        if (!confirm('Вы уверены, что хотите удалить этот комментарий и все ответы на него?')) {
            return;
        }

        try {
            const response = await fetch(`/comments/${commentId}`, {
                method: 'DELETE'
            });

            if (!response.ok) throw new Error('Ошибка при удалении комментария');

            this.showSuccess('Комментарий удален!');
            await this.loadComments();

        } catch (error) {
            this.showError('Ошибка при удалении: ' + error.message);
        }
    }

    handleSearch() {
        this.searchQuery = document.getElementById('searchInput')?.value.trim() || '';
        this.currentPage = 1;
        this.loadComments();
    }

    clearSearch() {
        document.getElementById('searchInput').value = '';
        this.searchQuery = '';
        this.currentPage = 1;
        this.loadComments();
    }

    handleSortChange() {
        this.sortBy = document.getElementById('sortBy')?.value || 'created_at';
        this.sortOrder = document.getElementById('sortOrder')?.value || 'desc';
        this.currentPage = 1;
        this.loadComments();
    }

    prevPage() {
        if (this.currentPage > 1) {
            this.currentPage--;
            this.loadComments();
        }
    }

    nextPage() {
        this.currentPage++;
        this.loadComments();
    }

    updatePagination(data) {
        document.getElementById('prevPage').disabled = !data.has_prev;
        document.getElementById('nextPage').disabled = !data.has_next;
        document.getElementById('pageInfo').textContent = `Страница ${data.page} из ${Math.ceil(data.total / data.page_size)}`;
    }

    delegateEvents() {
        // Обработка кликов на комментарии
        document.getElementById('commentsContainer')?.addEventListener('click', (e) => {
            const replyBtn = e.target.closest('.reply-btn');
            const deleteBtn = e.target.closest('.delete-btn');
            
            if (replyBtn) {
                const commentId = parseInt(replyBtn.dataset.commentId);
                const authorName = replyBtn.dataset.author;
                this.replyTo(commentId, authorName);
            }
            
            if (deleteBtn) {
                const commentId = parseInt(deleteBtn.dataset.commentId);
                this.deleteComment(commentId);
            }
        });
    }

    showError(message) {
        this.showMessage(message, 'error');
    }

    showSuccess(message) {
        this.showMessage(message, 'success');
    }

    showMessage(message, type) {
        const messageEl = document.createElement('div');
        messageEl.className = type;
        messageEl.innerHTML = `
            <i class="fas fa-${type === 'error' ? 'exclamation-circle' : 'check-circle'}"></i>
            ${this.escapeHtml(message)}
        `;

        const container = document.querySelector('.container');
        if (container) {
            container.insertBefore(messageEl, container.firstChild);
            
            setTimeout(() => messageEl.remove(), 5000);
        }
    }

    escapeHtml(unsafe) {
        if (!unsafe) return '';
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
}

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
    const app = new CommentsApp();
    app.delegateEvents();
    window.app = app;
});