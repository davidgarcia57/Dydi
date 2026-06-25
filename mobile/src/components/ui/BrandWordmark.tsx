import React from 'react';
import { View, Text } from 'react-native';
import DydiLogo from './DydiLogo';

interface BrandWordmarkProps {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  showText?: boolean;
}

const TEXT_SIZES = {
  sm: 'text-xl',
  md: 'text-2xl',
  lg: 'text-4xl',
  xl: 'text-5xl',
};

const LOGO_SIZES = {
  sm: 24,
  md: 32,
  lg: 48,
  xl: 64,
};

export default function BrandWordmark({ size = 'md', showText = true }: BrandWordmarkProps) {
  return (
    <View className="flex-row items-center gap-2">
      <DydiLogo size={LOGO_SIZES[size]} />
      {showText && (
        <Text className={`font-serif text-terracotta tracking-tight ${TEXT_SIZES[size]}`}>
          DYDI
        </Text>
      )}
    </View>
  );
}
