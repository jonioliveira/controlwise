class CreateItems < ActiveRecord::Migration[7.1]
  def change
    create_table :items do |t|
      t.string :name
      t.string :description
      t.decimal :price, precision: 10, scale: 2
      t.decimal :margin, precision: 10, scale: 2
      t.integer :vat
      t.references :budget, null: false, foreign_key: true

      t.timestamps
    end
  end
end
