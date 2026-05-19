ALTER TABLE user_company_memberships
ADD COLUMN price_visibility JSONB NOT NULL DEFAULT '{
  "material_price": false,
  "labor_cost": false,
  "profit": false
}';

-- admin และ manager เห็นทุกอย่างโดย default
UPDATE user_company_memberships
SET price_visibility = '{"material_price": true, "labor_cost": true, "profit": true}'
WHERE role IN ('admin', 'manager');
