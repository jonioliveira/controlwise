# frozen_string_literal: true

require "pagy/extras/bootstrap"
require "pagy/extras/countless"

Pagy::DEFAULT[:items] = 15
Pagy::DEFAULT[:size] = []
Pagy::DEFAULT.freeze
