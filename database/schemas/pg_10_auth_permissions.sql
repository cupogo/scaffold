-- auth permission data
INSERT INTO auth_permission ("id", "creator_id", "updated", "name", "path", "method", "is_active", "remark") VALUES 
('v1-accounts-get', 0, CURRENT_TIMESTAMP, 'API: 列出账号 🔑', '/api/v1/accounts', 'GET', true, ''),
('v1-accounts-id-delete', 0, CURRENT_TIMESTAMP, 'API: 删除账号 🔑', '/api/v1/accounts/{id}', 'DELETE', true, ''),
('v1-accounts-id-get', 0, CURRENT_TIMESTAMP, 'API: 获取账号 🔑', '/api/v1/accounts/{id}', 'GET', true, ''),
('v1-accounts-id-put', 0, CURRENT_TIMESTAMP, 'API: 更新账号 🔑', '/api/v1/accounts/{id}', 'PUT', true, ''),
('v1-accounts-post', 0, CURRENT_TIMESTAMP, 'API: 录入账号 🔑', '/api/v1/accounts', 'POST', true, ''),
('v1-cms-articles-id-delete', 0, CURRENT_TIMESTAMP, 'API: 删除文章 🔑', '/api/v1/cms/articles/{id}', 'DELETE', true, ''),
('v1-cms-articles-id-put', 0, CURRENT_TIMESTAMP, 'API: 更新文章 🔑', '/api/v1/cms/articles/{id}', 'PUT', true, ''),
('v1-cms-articles-post', 0, CURRENT_TIMESTAMP, 'API: 录入文章 🔑', '/api/v1/cms/articles', 'POST', true, ''),
('v1-cms-attachments-id-delete', 0, CURRENT_TIMESTAMP, 'API: 删除附件 🔑', '/api/v1/cms/attachments/{id}', 'DELETE', true, ''),
('v1-cms-attachments-post', 0, CURRENT_TIMESTAMP, 'API: 录入附件 🔑', '/api/v1/cms/attachments', 'POST', true, ''),
('v1-cms-clauses-id-delete', 0, CURRENT_TIMESTAMP, 'API: 删除内容条款 🔑', '/api/v1/cms/clauses/{id}', 'DELETE', true, ''),
('v1-cms-clauses-id-put', 0, CURRENT_TIMESTAMP, 'API: 录入内容条款 🔑', '/api/v1/cms/clauses/{id}', 'PUT', true, '')
ON CONFLICT (id) DO UPDATE SET updated = CURRENT_TIMESTAMP;

DELETE FROM auth_permission WHERE id SIMILAR TO 'v1-%' AND updated < CURRENT_DATE -1;
