# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[7.1].define(version: 2024_06_25_110847) do
  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "budgets", force: :cascade do |t|
    t.string "name"
    t.string "description"
    t.decimal "total", precision: 10, scale: 2
    t.decimal "margin", precision: 10, scale: 2
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  create_table "budgets_items", id: false, force: :cascade do |t|
    t.bigint "budget_id", null: false
    t.bigint "item_id", null: false
  end

  create_table "items", force: :cascade do |t|
    t.string "name"
    t.string "description"
    t.decimal "price", precision: 10, scale: 2
    t.decimal "margin", precision: 10, scale: 2
    t.integer "vat"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "unit"
  end

  create_table "users", force: :cascade do |t|
    t.string "email", default: "", null: false
    t.string "encrypted_password", default: "", null: false
    t.string "reset_password_token"
    t.datetime "reset_password_sent_at"
    t.datetime "remember_created_at"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["email"], name: "index_users_on_email", unique: true
    t.index ["reset_password_token"], name: "index_users_on_reset_password_token", unique: true
  end

  create_table "versions", id: :uuid, default: -> { "gen_random_uuid()" }, force: :cascade do |t|
    t.string "item_type", null: false
    t.string "item_id", null: false
    t.string "event", null: false
    t.string "whodunnit"
    t.text "object"
    t.datetime "created_at"
    t.text "object_changes"
    t.index ["item_type", "item_id"], name: "index_versions_on_item_type_and_item_id"
  end

end
