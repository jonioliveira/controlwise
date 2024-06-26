class ItemsController < ApplicationController
  before_action :set_item, only: %i[ show edit update destroy ]

  # GET /items or /items.json
  def index
    @pagy, @items = pagy(Item.all)
  end

  def search
    @q = Item.ransack(params[:q])
    @items = @q.result(distinct: true)
    respond_to do |format|
      format.js { render partial: "search_results"}
    end
  end

  # GET /items/1 or /items/1.json
  def show
    @item = Item.find(params[:id])
    @versions = @item.versions.reject { |version| version.event == 'create' }.map do |version|
      changes = begin
                  YAML.safe_load(version.object_changes, permitted_classes: [
                    BigDecimal,
                    ActiveSupport::TimeWithZone,
                    ActiveSupport::TimeZone,
                    Time
                  ], aliases: true)
                rescue Psych::DisallowedClass
                  {}
                end

      # Exclude `updated_at` from the changes
      changes.except!('updated_at')

      # Translate the keys
      translated_changes = changes.transform_keys { |key| I18n.t("activerecord.attributes.item.#{key}") }

      { version: version, changes: translated_changes }
    end.reject { |entry| entry[:changes].empty? }
  end

  # GET /items/new
  def new
    @item = Item.new
  end

  # GET /items/1/edit
  def edit
  end

  # POST /items or /items.json
  def create
    @item = Item.new(item_params)

    respond_to do |format|
      if @item.save
        format.html { redirect_to item_url(@item), notice: "Item was successfully created." }
        format.json { render :show, status: :created, location: @item }
      else
        format.html { render :new, status: :unprocessable_entity }
        format.json { render json: @item.errors, status: :unprocessable_entity }
      end
    end
  end

  # PATCH/PUT /items/1 or /items/1.json
  def update
    respond_to do |format|
      if @item.update(item_params)
        format.html { redirect_to item_url(@item), notice: "Item was successfully updated." }
        format.json { render :show, status: :ok, location: @item }
      else
        format.html { render :edit, status: :unprocessable_entity }
        format.json { render json: @item.errors, status: :unprocessable_entity }
      end
    end
  end

  # DELETE /items/1 or /items/1.json
  def destroy
    @item.destroy!

    respond_to do |format|
      format.html { redirect_to items_url, notice: "Item was successfully destroyed." }
      format.json { head :no_content }
    end
  end

  private
    # Use callbacks to share common setup or constraints between actions.
    def set_item
      @item = Item.find(params[:id])
    end

    # Only allow a list of trusted parameters through.
    def item_params
      params.require(:item).permit(:name, :unit, :description, :price, :margin, :vat)
    end
end
