class CreateJoinTableBudgetItems < ActiveRecord::Migration[7.1]
  def change
    create_join_table :budgets, :items do |t|
      # t.index [:budget_id, :item_id]
      # t.index [:item_id, :budget_id]
    end
  end
end
