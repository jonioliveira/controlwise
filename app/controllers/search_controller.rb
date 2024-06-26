class SearchController < ApplicationController

  def index
  end

  def search
    @query = params[:query]
  end

end
