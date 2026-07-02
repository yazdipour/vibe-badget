INSERT OR IGNORE INTO categories (name, icon, color, icon_color, kind)
SELECT name, icon, color, icon_color, kind
FROM (
  SELECT 'Groceries' name, 'ShoppingCart' icon, '#4CAF50' color, '#ffffff' icon_color, 'expense' kind UNION ALL
  SELECT 'Eating Out', 'Utensils', '#FF9800', '#1f2937', 'expense' UNION ALL
  SELECT 'Transport', 'Car', '#2196F3', '#ffffff', 'expense' UNION ALL
  SELECT 'Shopping', 'Shirt', '#E91E63', '#ffffff', 'expense' UNION ALL
  SELECT 'Bills & Utilities', 'Zap', '#FFC107', '#1f2937', 'expense' UNION ALL
  SELECT 'Salary', 'Wallet', '#2E7D32', '#ffffff', 'income' UNION ALL
  SELECT 'Savings', 'PiggyBank', '#009688', '#ffffff', 'income' UNION ALL
  SELECT 'Entertainment', 'Gamepad2', '#9C27B0', '#ffffff', 'expense' UNION ALL
  SELECT 'Health', 'HeartPulse', '#F44336', '#ffffff', 'expense' UNION ALL
  SELECT 'Uncategorized', 'Tag', '#607D8B', '#ffffff', 'expense' UNION ALL
  SELECT 'Ignore', 'Tag', '#9E9E9E', '#1f2937', 'expense'
)
WHERE NOT EXISTS (SELECT 1 FROM settings WHERE key='default_seed_v2');

-- Groceries rules
INSERT OR IGNORE INTO rules (field, match_type, pattern, category_id)
  SELECT 'partner_name', 'keyword', 'Lidl', id FROM categories
  WHERE name='Groceries' AND NOT EXISTS (SELECT 1 FROM settings WHERE key='default_seed_v2');

INSERT OR IGNORE INTO settings (key, value) VALUES ('default_seed_v2', 'true');
