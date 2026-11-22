class Launchbox < Formula
  # Homebrew formulas values are hardcoded and the reason
  # is due to how Homebrew’s DSL (domain-specific language in Ruby)
  # and its formula system are designed and parsed
  desc "Command-line interface for creating boilerplate Next.js application."
  homepage "https://github.com/fabiomrosa/homebrew-launchbox"
  license "MIT"
  version "0.7.0"

  on_macos do
    on_arm do
      url "https://github.com/fabiomrosa/homebrew-launchbox/releases/download/v#{version}/launchbox-darwin-arm64"
      sha256 "bbea16fca1cc9598edd2da982dd217b42302fc9f1ab86745f90f5a687ed61747"
    end

    on_intel do
      url "https://github.com/fabiomrosa/homebrew-launchbox/releases/download/v#{version}/launchbox-darwin-amd64"
      sha256 "be3fa0f0a91498760df697612aea9426a41e466b4a283951481a8232153da44a"
    end
  end

  def install
    binary = if Hardware::CPU.arm?
      "launchbox-darwin-arm64"
    else
      "launchbox-darwin-amd64"
    end

    bin.install binary => "launchbox"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/launchbox --version")
  end
end
