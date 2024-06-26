class Item < ApplicationRecord
  has_paper_trail skip: [:updated_at]
  has_many :budget_items
  has_many :item, through: :budget_items

  validates :name, presence: true
  validates :description, presence: true
  validates :quantity, numericality: { greater_than_or_equal_to: 0 }
  validates :margin, numericality: { greater_than_or_equal_to: 0 }

  def self.ransackable_attributes(auth_object = nil)
    # Specify the attributes you want to allow for searching
    %w[name description]
  end
end
