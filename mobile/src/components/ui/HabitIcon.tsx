import React from 'react';
import { View } from 'react-native';
import Svg, { Path, Circle, Rect, Line, G, Ellipse } from 'react-native-svg';

interface HabitIconProps {
  iconKey?: string;
  size?: number;
  color?: string;
}

export default function HabitIcon({ iconKey = '', size = 40, color = '#2A251F' }: HabitIconProps) {
  // Static implementations of the SVGs
  const renderIcon = () => {
    switch (iconKey) {
      case 'exercise':
        return (
          <G>
            <Line x1="8" y1="12" x2="16" y2="12" />
            <Rect x="4.5" y="7.5" width="3" height="9" rx="1.2" />
            <Rect x="16.5" y="7.5" width="3" height="9" rx="1.2" />
            <Line x1="3" y1="9.5" x2="3" y2="14.5" />
            <Line x1="21" y1="9.5" x2="21" y2="14.5" />
          </G>
        );
      case 'water':
        return (
          <G>
            <Path d="M12 3.5 C12 3.5 6 10.5 6 14.8 A6 6 0 0 0 18 14.8 C18 10.5 12 3.5 12 3.5 Z" />
            <Path d="M9.4 15.2 A2.8 2.8 0 0 0 11.6 17.4" opacity={0.6} />
          </G>
        );
      case 'steps':
        return (
          <G>
            <G>
              <Ellipse cx="8.5" cy="9" rx="2.2" ry="3.2" />
              <Circle cx="8.5" cy="13.6" r="1.1" />
            </G>
            <G>
              <Ellipse cx="15.5" cy="13" rx="2.2" ry="3.2" />
              <Circle cx="15.5" cy="17.6" r="1.1" />
            </G>
          </G>
        );
      case 'fruit':
        return (
          <G>
            <Circle cx="12" cy="13.5" r="5.5" />
            <Path d="M12 8 C12.6 5.8 14.6 5.2 16.2 5.4 C15.9 7.4 14.2 8.4 12 8 Z" />
          </G>
        );
      case 'no_sugar':
        return (
          <G>
            <Path d="M7.5 8 L16.5 8 L15.6 19 Q15.5 20 14.5 20 L9.5 20 Q8.5 20 8.4 19 Z" />
            <Line x1="6" y1="8" x2="18" y2="8" />
            <Line x1="13" y1="4.2" x2="11.6" y2="8" />
            <Line x1="5" y1="19" x2="19" y2="5" opacity={0.5} />
          </G>
        );
      case 'read':
        return (
          <G>
            <Path d="M12 7 C10 5.8 7 5.6 4.5 6.1 L4.5 16.8 C7 16.3 10 16.6 12 17.9" />
            <Path d="M12 7 C14 5.8 17 5.6 19.5 6.1 L19.5 16.8 C17 16.3 14 16.6 12 17.9" />
            <Line x1="12" y1="7" x2="12" y2="17.9" />
          </G>
        );
      case 'focus':
        return (
          <G>
            <Circle cx="12" cy="12" r="8" opacity={0.5} />
            <Circle cx="12" cy="12" r="4.4" />
            <Circle cx="12" cy="12" r="1.4" fill={color} stroke="none" />
          </G>
        );
      case 'journal':
        return (
          <G>
            <Rect x="6" y="4" width="12" height="16" rx="1.6" />
            <Line x1="9" y1="8.5" x2="15" y2="8.5" />
            <Line x1="9" y1="12" x2="15" y2="12" />
            <Line x1="9" y1="15.5" x2="14" y2="15.5" />
          </G>
        );
      case 'no_social':
        return (
          <G>
            <Path d="M5 6 H19 A1.6 1.6 0 0 1 20.5 7.6 V13.4 A1.6 1.6 0 0 1 19 15 H12 L8 18 V15 H5 A1.6 1.6 0 0 1 3.5 13.4 V7.6 A1.6 1.6 0 0 1 5 6 Z" />
            <Line x1="4" y1="19.5" x2="20" y2="4.5" opacity={0.5} />
          </G>
        );
      case 'no_phone':
        return (
          <G>
            <Rect x="7" y="3" width="10" height="18" rx="2.5" />
            <Line x1="10.5" y1="18.2" x2="13.5" y2="18.2" />
            <Path d="M14.6 9 A3 3 0 1 1 11.4 6 A2.3 2.3 0 0 0 14.6 9 Z" fill={color} stroke="none" />
          </G>
        );
      case 'bed':
        return (
          <G>
            <Path d="M4 17 V12 A2 2 0 0 1 6 10 H18 A2 2 0 0 1 20 12 V17" />
            <Line x1="3" y1="17" x2="21" y2="17" />
            <Line x1="3.5" y1="17" x2="3.5" y2="19" />
            <Line x1="20.5" y1="17" x2="20.5" y2="19" />
            <Path d="M7 10 V8.4 A1.4 1.4 0 0 1 8.4 7 H10.6 A1.4 1.4 0 0 1 12 8.4 V10" />
            <Path d="M16.6 6 L17.1 7.3 L18.4 7.8 L17.1 8.3 L16.6 9.6 L16.1 8.3 L14.8 7.8 L16.1 7.3 Z" fill={color} stroke="none" />
          </G>
        );
      case 'outdoors':
        return (
          <G>
            <Circle cx="12" cy="12" r="3.8" />
            <G>
              <Line x1="12" y1="2.5" x2="12" y2="5" />
              <Line x1="12" y1="19" x2="12" y2="21.5" />
              <Line x1="2.5" y1="12" x2="5" y2="12" />
              <Line x1="19" y1="12" x2="21.5" y2="12" />
              <Line x1="5.2" y1="5.2" x2="6.9" y2="6.9" />
              <Line x1="17.1" y1="17.1" x2="18.8" y2="18.8" />
              <Line x1="5.2" y1="18.8" x2="6.9" y2="17.1" />
              <Line x1="17.1" y1="6.9" x2="18.8" y2="5.2" />
            </G>
          </G>
        );
      default:
        return (
          <Path d="M12 4 L13.3 10.7 L20 12 L13.3 13.3 L12 20 L10.7 13.3 L4 12 L10.7 10.7 Z" />
        );
    }
  };

  return (
    <View style={{ width: size, height: size }}>
      <Svg
        width="100%"
        height="100%"
        viewBox="0 0 24 24"
        fill="none"
        stroke={color}
        strokeWidth="1.7"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        {renderIcon()}
      </Svg>
    </View>
  );
}
