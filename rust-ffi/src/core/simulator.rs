use anyhow::Result;
use enigo::{Enigo, Key, Keyboard, Settings};
use tracing::debug;

pub struct KeySimulator {
    enigo: Enigo,
}

impl KeySimulator {
    pub fn new() -> Result<Self> {
        Ok(Self {
            enigo: Enigo::new(&Settings::default())
                .map_err(|e| anyhow::anyhow!("Failed to initialize key simulator: {}", e))?,
        })
    }

    pub fn select_all(&mut self) -> Result<()> {
        debug!("Simulating select all");

        #[cfg(target_os = "macos")]
        {
            self.enigo.key(Key::Meta, enigo::Direction::Press)?;
            self.enigo.key(Key::Unicode('a'), enigo::Direction::Click)?;
            self.enigo.key(Key::Meta, enigo::Direction::Release)?;
        }

        #[cfg(not(target_os = "macos"))]
        {
            self.enigo.key(Key::Control, enigo::Direction::Press)?;
            self.enigo.key(Key::Unicode('a'), enigo::Direction::Click)?;
            self.enigo.key(Key::Control, enigo::Direction::Release)?;
        }

        Ok(())
    }

    pub fn copy(&mut self) -> Result<()> {
        debug!("Simulating copy");

        #[cfg(target_os = "macos")]
        {
            self.enigo.key(Key::Meta, enigo::Direction::Press)?;
            self.enigo.key(Key::Unicode('c'), enigo::Direction::Click)?;
            self.enigo.key(Key::Meta, enigo::Direction::Release)?;
        }

        #[cfg(not(target_os = "macos"))]
        {
            self.enigo.key(Key::Control, enigo::Direction::Press)?;
            self.enigo.key(Key::Unicode('c'), enigo::Direction::Click)?;
            self.enigo.key(Key::Control, enigo::Direction::Release)?;
        }

        Ok(())
    }

    pub fn paste(&mut self) -> Result<()> {
        debug!("Simulating paste");

        #[cfg(target_os = "macos")]
        {
            self.enigo.key(Key::Meta, enigo::Direction::Press)?;
            self.enigo.key(Key::Unicode('v'), enigo::Direction::Click)?;
            self.enigo.key(Key::Meta, enigo::Direction::Release)?;
        }

        #[cfg(not(target_os = "macos"))]
        {
            self.enigo.key(Key::Control, enigo::Direction::Press)?;
            self.enigo.key(Key::Unicode('v'), enigo::Direction::Click)?;
            self.enigo.key(Key::Control, enigo::Direction::Release)?;
        }

        Ok(())
    }
}
