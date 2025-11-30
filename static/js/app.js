class CommentsApp {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 10;
        this.searchQuery = '';
        this.sortBy = 'created_at';
        this.sortOrder = 'desc';
        this.replyingTo = null;
        this.selectedPost = null;
        this.posts = [];
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadPosts();
    }

    bindEvents() {
        // Создание поста
        document.getElementById('createPostBtn').addEventListener('click', () => this.createPost());
        
        // Закрытие выбранного поста
        document.getElementById('closePost').addEventListener('click', () => this.closeSelectedPost());

        // Форма комментария
        document.getElementById('commentForm').addEventListener('submit', (e) => this.handleSubmit(e));
        document.getElementById('cancelReply').addEventListener('click', () => this.cancelReply());

        // Поиск
        document.getElementById('searchBtn').addEventListener('click', () => this.handleSearch());
        document.getElementById('searchInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.handleSearch();
        });
        document.getElementById('clearSearch').addEventListener('click', () => this.clearSearch());

        // Сортировка
        document.getElementById('sortBy').addEventListener('change', () => this.handleSortChange());
        document.getElementById('sortOrder').addEventListener('change', () => this.handleSortChange());

        // Пагинация
        document.getElementById('prevPage').addEventListener('click', () => this.prevPage());
        document.getElementById('nextPage').addEventListener('click', () => this.nextPage());

        // Модальное окно
        document.getElementById('confirmDelete').addEventListener('click', () => this.confirmDelete());
        document.getElementById('cancelDelete').addEventListener('click', () => this.hideDeleteModal());
    }

    async loadPosts() {
        const list = document.getElementById('postsList');
        list.innerHTML = '<div class="loading"><div class="loading-spinner"></div> Загрузка постов...</div>';

        try {
            const response = await fetch('/posts?page=1&page_size=100');
            if (!response.ok) throw new Error('Ошибка загрузки постов');
            
            const data = await response.json();
            this.posts = data.posts;
            this.renderPosts();
        } catch (error) {
            this.showError('Ошибка при загрузке постов: ' + error.message);
        }
    }

    renderPosts() {
        const list = document.getElementById('postsList');
        
        if (this.posts.length === 0) {
            list.innerHTML = `
                <div class="loading">
                    <i class="fas fa-inbox" style="font-size: 3rem; margin-bottom: 1rem; color: #bdc3c7;"></i>
                    <p>Постов пока нет. Создайте первый пост!</p>
                </div>
            `;
            return;
        }

        list.innerHTML = this.posts.map(post => `
            <div class="post-item" onclick="app.selectPost(${post.id})">
                <div class="post-header">
                    <div class="post-title">${this.escapeHtml(post.title)}</div>
                    <div class="post-meta">
                        <span class="author"><i class="fas fa-user"></i> ${this.escapeHtml(post.author)}</span>
                        <span class="post-date"><i class="far fa-clock"></i> ${new Date(post.created_at).toLocaleString('ru-RU')}</span>
                    </div>
                </div>
                <div class="post-content">${this.escapeHtml(post.content)}</div>
                <div class="post-stats">
                    <span><i class="fas fa-comments"></i> ${post.comments_count || 0} комментариев</span>
                    <span><i class="far fa-eye"></i> Открыть комментарии</span>
                </div>
            </div>
        `).join('');
    }

    selectPost(postId) {
        const post = this.posts.find(p => p.id === postId);
        if (!post) return;

        this.selectedPost = post;
        this.showSelectedPost();
        this.loadComments();
    }

    showSelectedPost() {
        document.getElementById('selectedPostSection').style.display = 'block';
        document.getElementById('selectedPostTitle').textContent = this.escapeHtml(this.selectedPost.title);
        document.getElementById('selectedPostAuthor').textContent = this.escapeHtml(this.selectedPost.author);
        document.getElementById('selectedPostDate').textContent = new Date(this.selectedPost.created_at).toLocaleString('ru-RU');
        document.getElementById('selectedPostContent').textContent = this.escapeHtml(this.selectedPost.content);
        document.getElementById('selectedPostCommentsCount').textContent = this.selectedPost.comments_count || 0;

        // Прокрутка к выбранному посту
        document.getElementById('selectedPostSection').scrollIntoView({ behavior: 'smooth' });
    }

    closeSelectedPost() {
        this.selectedPost = null;
        document.getElementById('selectedPostSection').style.display = 'none';
        this.replyingTo = null;
        this.cancelReply();
    }

    async createPost() {
        const title = document.getElementById('postTitle').value.trim();
        const content = document.getElementById('postContent').value.trim();
        const author = document.getElementById('postAuthor').value.trim();

        if (!title || !content || !author) {
            this.showError('Пожалуйста, заполните все поля поста');
            return;
        }

        const postData = {
            title: title,
            content: content,
            author: author
        };

        try {
            const response = await fetch('/posts', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(postData)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Ошибка при создании поста');
            }

            this.showSuccess('Пост успешно создан!');
            
            // Очистка формы
            document.getElementById('postTitle').value = '';
            document.getElementById('postContent').value = '';
            document.getElementById('postAuthor').value = '';
            
            // Перезагрузка списка постов
            this.loadPosts();

        } catch (error) {
            this.showError('Ошибка: ' + error.message);
        }
    }

    async loadComments() {
        if (!this.selectedPost) return;

        const list = document.getElementById('commentsList');
        list.innerHTML = '<div class="loading"><div class="loading-spinner"></div> Загрузка комментариев...</div>';

        try {
            const params = new URLSearchParams({
                post_id: this.selectedPost.id,
                page: this.currentPage,
                page_size: this.pageSize,
                search: this.searchQuery,
                sort_by: this.sortBy,
                sort_order: this.sortOrder
            });

            const response = await fetch(`/comments?${params}`);
            if (!response.ok) throw new Error('Ошибка загрузки комментариев');
            
            const data = await response.json();
            this.renderComments(data.comments);
            this.updatePagination(data);
            this.updateStats(data.total);

        } catch (error) {
            this.showError('Ошибка при загрузке комментариев: ' + error.message);
        }
    }

    renderComments(comments, level = 0) {
        const list = document.getElementById('commentsList');
        
        if (comments.length === 0) {
            list.innerHTML = `
                <div class="loading">
                    <i class="fas fa-comments" style="font-size: 3rem; margin-bottom: 1rem; color: #bdc3c7;"></i>
                    <p>Комментариев пока нет. Будьте первым!</p>
                </div>
            `;
            return;
        }

        list.innerHTML = comments.map(comment => this.renderComment(comment, level)).join('');
    }

    renderComment(comment, level = 0) {
        const date = new Date(comment.created_at).toLocaleString('ru-RU');
        const levelClass = level > 0 ? `comment-level-${Math.min(level, 5)}` : '';
        
        return `
            <div class="comment ${levelClass} ${this.replyingTo === comment.id ? 'replying' : ''}" data-comment-id="${comment.id}">
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
                    <button class="btn btn-outline btn-sm" onclick="app.replyTo(${comment.id}, '${this.escapeHtml(comment.author)}')">
                        <i class="fas fa-reply"></i> Ответить
                    </button>
                    <button class="btn btn-danger btn-sm" onclick="app.showDeleteModal(${comment.id})">
                        <i class="fas fa-trash"></i> Удалить
                    </button>
                </div>
                ${comment.children && comment.children.length > 0 ? 
                    `<div class="comment-children">
                        ${comment.children.map(child => this.renderComment(child, level + 1)).join('')}
                    </div>` : ''
                }
            </div>
        `;
    }

    async handleSubmit(e) {
        e.preventDefault();
        
        if (!this.selectedPost) {
            this.showError('Пожалуйста, выберите пост для комментария');
            return;
        }
        
        const author = document.getElementById('commentAuthor').value.trim();
        const content = document.getElementById('commentContent').value.trim();
        const parentId = document.getElementById('parentId').value;

        if (!author || !content) {
            this.showError('Пожалуйста, заполните все поля');
            return;
        }

        const commentData = {
            post_id: this.selectedPost.id,
            author: author,
            content: content,
            parent_id: parentId ? parseInt(parentId) : null
        };

        try {
            const response = await fetch('/comments', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(commentData)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Ошибка при создании комментария');
            }

            this.showSuccess('Комментарий успешно добавлен!');
            
            // Очистка формы
            document.getElementById('commentForm').reset();
            this.cancelReply();
            
            // Перезагрузка комментариев и постов
            this.loadComments();
            this.loadPosts();

        } catch (error) {
            this.showError('Ошибка: ' + error.message);
        }
    }

    replyTo(commentId, authorName) {
        this.replyingTo = commentId;
        document.getElementById('parentId').value = commentId;
        document.getElementById('cancelReply').style.display = 'inline-block';
        
        const contentInput = document.getElementById('commentContent');
        contentInput.placeholder = `Ответ ${authorName}...`;
        contentInput.focus();

        // Прокрутка к форме
        document.querySelector('.comment-form').scrollIntoView({ behavior: 'smooth' });
    }

    cancelReply() {
        this.replyingTo = null;
        document.getElementById('parentId').value = '';
        document.getElementById('cancelReply').style.display = 'none';
        document.getElementById('commentContent').placeholder = 'Ваш комментарий...';
    }

    async deleteComment(commentId) {
        try {
            const response = await fetch(`/comments/${commentId}`, {
                method: 'DELETE'
            });

            if (!response.ok) throw new Error('Ошибка при удалении комментария');

            this.showSuccess('Комментарий удален!');
            this.loadComments();
            this.loadPosts(); // Обновляем счетчик комментариев

        } catch (error) {
            this.showError('Ошибка при удалении: ' + error.message);
        }
    }

    handleSearch() {
        this.searchQuery = document.getElementById('searchInput').value.trim();
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
        this.sortBy = document.getElementById('sortBy').value;
        this.sortOrder = document.getElementById('sortOrder').value;
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

    updateStats(total) {
        const stats = document.getElementById('commentsStats');
        stats.textContent = `Всего комментариев: ${total}`;
    }

    showDeleteModal(commentId) {
        this.commentToDelete = commentId;
        document.getElementById('deleteModal').style.display = 'flex';
    }

    hideDeleteModal() {
        this.commentToDelete = null;
        document.getElementById('deleteModal').style.display = 'none';
    }

    confirmDelete() {
        if (this.commentToDelete) {
            this.deleteComment(this.commentToDelete);
            this.hideDeleteModal();
        }
    }

    showError(message) {
        this.showMessage(message, 'error');
    }

    showSuccess(message) {
        this.showMessage(message, 'success');
    }

    showMessage(message, type) {
        // Удаляем старые сообщения
        const oldMessages = document.querySelectorAll('.error, .success');
        oldMessages.forEach(msg => msg.remove());

        const messageEl = document.createElement('div');
        messageEl.className = type;
        messageEl.textContent = message;

        // Вставляем сообщение в начало контейнера
        const container = document.querySelector('.container');
        container.insertBefore(messageEl, container.firstChild);

        // Автоудаление через 5 секунд
        setTimeout(() => {
            if (messageEl.parentNode) {
                messageEl.remove();
            }
        }, 5000);
    }

    escapeHtml(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
}

// Инициализация приложения
const app = new CommentsApp();