import React, { useEffect, useRef } from 'react';
import { Animated, Easing } from 'react-native';
import Svg, { G, Circle } from 'react-native-svg';

interface DydiLogoProps {
  size?: number;
}

const AnimatedG = Animated.createAnimatedComponent(G);

export default function DydiLogo({ size = 32 }: DydiLogoProps) {
  const spinValue = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    const spinAnimation = Animated.loop(
      Animated.timing(spinValue, {
        toValue: 1,
        duration: 30000, // 30s rotation like web
        easing: Easing.linear,
        useNativeDriver: true,
      })
    );
    spinAnimation.start();
    return () => spinAnimation.stop();
  }, [spinValue]);

  const spin = spinValue.interpolate({
    inputRange: [0, 1],
    outputRange: ['0deg', '360deg'],
  });

  // Scale calculations
  const scale = size / 240;

  return (
    <Svg width={size} height={size} viewBox="0 0 240 240">
      <G transform="translate(120, 120)">
        <AnimatedG style={{ transform: [{ rotate: spin }] }}>
          {/* Segment 3: Hairline */}
          <Circle
            cx="0"
            cy="0"
            r="70"
            fill="none"
            stroke="#E7DECD" // hairline
            strokeWidth="20"
            strokeLinecap="round"
            strokeDasharray="115 325"
            transform="rotate(240)"
          />
          {/* Segment 1: Terracotta */}
          <Circle
            cx="0"
            cy="0"
            r="70"
            fill="none"
            stroke="#C26F4D" // terracotta
            strokeWidth="20"
            strokeLinecap="round"
            strokeDasharray="115 325"
            transform="rotate(0)"
          />
          {/* Segment 2: Sage-deep (Displaced) */}
          <G transform="rotate(120) translate(0, 12)">
            <Circle
              cx="0"
              cy="0"
              r="70"
              fill="none"
              stroke="#7CA39D" // sage-deep
              strokeWidth="20"
              strokeLinecap="round"
              strokeDasharray="115 325"
            />
          </G>
          {/* Roulette balls */}
          <Circle cx="82" cy="-40" r="7" fill="#2A251F" />
          <Circle cx="94" cy="-22" r="4" fill="#6F6557" />
        </AnimatedG>
      </G>
    </Svg>
  );
}
