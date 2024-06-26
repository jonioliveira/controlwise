class ApplicationController < ActionController::Base
  include Pagy::Backend
  before_action :authenticate_user!
  before_action :set_paper_trail_whodunnit

  private

  def user_for_paper_trail
    user_signed_in? ? current_user.email : 'Unknown user'
  end
end
