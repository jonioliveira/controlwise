class Budget < ApplicationRecord
  has_paper_trail
  belongs_to :user
  has_many :budget_items
  has_many :item, through: :budget_items
  validates :name, presence: true
  validates :name, uniqueness: { scope: :user_id, message: "You already have a budget with that name" }
end
