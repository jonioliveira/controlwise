class BudgetItem < ApplicationRecord
  belongs_to :budget
  belongs_to :item
end
