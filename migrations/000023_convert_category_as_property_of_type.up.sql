ALTER TABLE projecta_cost_types ADD COLUMN category_id uuid REFERENCES projecta_cost_categories(category_id);

with
    t1 as (select projecta_expenses.project_id, type_id, category_id from projecta_expenses group by project_id, type_id, category_id)
update projecta_cost_types set category_id = t1.category_id from t1 where projecta_cost_types.type_id = t1.type_id and projecta_cost_types.project_id = t1.project_id;

ALTER TABLE projecta_cost_types ALTER COLUMN category_id SET NOT NULL;

ALTER TABLE projecta_expenses DROP COLUMN category_id;
