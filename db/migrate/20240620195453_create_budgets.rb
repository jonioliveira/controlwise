class CreateBudgets < ActiveRecord::Migration[7.1]
  def change
    create_table :budgets do |t|
      t.string :name
      t.string :description
      t.decimal :total, precision: 10, scale: 2
      t.decimal :margin, precision: 10, scale: 2

      t.timestamps
    end
  end
end
