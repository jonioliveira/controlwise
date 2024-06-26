class AddUnitToItem < ActiveRecord::Migration[7.1]
  def change
    add_column :items, :unit, :string
  end
end
