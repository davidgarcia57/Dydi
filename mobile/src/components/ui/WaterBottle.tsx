import React, { useEffect } from 'react';
import { TouchableWithoutFeedback } from 'react-native';
import Svg, { Defs, ClipPath, Path, Rect, Circle, G } from 'react-native-svg';
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withRepeat,
  withSequence,
  withSpring,
  withTiming,
} from 'react-native-reanimated';

interface WaterBottleProps {
  size?: number;
}

export default function WaterBottle({ size = 160 }: WaterBottleProps) {
  // El agua "ondula" con un balanceo sutil de la botella (loop de reanimated)
  // y chapotea al tocarla — espejo del WaterBottle animado del frontend web.
  const clipId = "wb-clip-rn";

  const rock = useSharedValue(0);
  useEffect(() => {
    rock.value = withRepeat(
      withSequence(
        withTiming(-1.6, { duration: 1900, easing: Easing.inOut(Easing.quad) }),
        withTiming(1.6, { duration: 1900, easing: Easing.inOut(Easing.quad) })
      ),
      -1
    );
  }, [rock]);

  const animStyle = useAnimatedStyle(() => ({
    transform: [{ rotate: `${rock.value}deg` }],
  }));

  function slosh() {
    rock.value = withSequence(
      withTiming(5, { duration: 130 }),
      withTiming(-4, { duration: 180 }),
      withSpring(0)
    );
  }

  return (
    <TouchableWithoutFeedback onPress={slosh}>
      <Animated.View style={[{ width: size, aspectRatio: 200 / 360 }, animStyle]}>
        <Svg width="100%" height="100%" viewBox="0 0 200 360">
          <Defs>
            <ClipPath id={clipId}>
              <Path d="M84 54 L84 73 Q84 80 79 84 Q59 94 59 118 L59 290 Q59 315 84 315 L116 315 Q141 315 141 290 L141 118 Q141 94 121 84 Q116 80 116 73 L116 54 Z" />
            </ClipPath>
          </Defs>

          <G clipPath={`url(#${clipId})`}>
            {/* Water */}
            <Rect x="-60" y="184" width="320" height="200" fill="#4F9FB0" />
            <Path
              d="M-100 184 q25 8 50 0 t50 0 t50 0 t50 0 t50 0 t50 0 t50 0 t50 0 L300 360 L-100 360 Z"
              fill="#4F9FB0"
            />
            <Path
              d="M-100 178 q25 -9 50 0 t50 0 t50 0 t50 0 t50 0 t50 0 t50 0 t50 0 L300 360 L-100 360 Z"
              fill="#7FC4D1"
            />
            {/* Bubbles */}
            <Circle cx="86" cy="300" r="3.5" fill="#ffffff" opacity="0.5" />
            <Circle cx="108" cy="305" r="2.5" fill="#ffffff" opacity="0.5" />
            <Circle cx="98" cy="298" r="3" fill="#ffffff" opacity="0.5" />
          </G>

          {/* Highlights and Outline */}
          <Rect x="66" y="120" width="7" height="150" rx="3.5" fill="#ffffff" opacity="0.22" />
          <Path
            d="M80 50 L80 72 Q80 79 74 83 Q54 93 54 117 L54 292 Q54 320 82 320 L118 320 Q146 320 146 292 L146 117 Q146 93 126 83 Q120 79 120 72 L120 50 Z"
            fill="none"
            stroke="#2A251F" /* ink */
            strokeWidth="4"
            strokeLinejoin="round"
          />
          {/* Cap */}
          <Rect x="77" y="20" width="46" height="30" rx="7" fill="#C26F4D" /* terracotta */ />
          <Rect x="82" y="44" width="36" height="9" rx="3" fill="#C26F4D" />
        </Svg>
      </Animated.View>
    </TouchableWithoutFeedback>
  );
}
