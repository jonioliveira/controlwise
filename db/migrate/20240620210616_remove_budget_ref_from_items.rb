class RemoveBudgetRefFromItems < ActiveRecord::Migration[7.1]
  def change
    remove_reference :items, :budget, null: false, foreign_key: true
  end
end
