#!/bin/bash

# Go Crypto Analyzer 启动脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 显示菜单
show_menu() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${GREEN}Go Crypto Analyzer${NC}"
    echo -e "${BLUE}================================${NC}"
    echo "1. 市场分析 - 查看当前市场状况"
    echo "2. 策略回测 - 测试交易策略"
    echo "3. 双向回测 - 支持做空的回测"
    echo "4. 构建项目"
    echo "5. 退出"
    echo -e "${BLUE}================================${NC}"
}

# 市场分析菜单
market_analysis() {
    echo -e "\n${GREEN}市场分析模块${NC}"
    echo "1. 分析BTC（默认）"
    echo "2. 分析ETH"
    echo "3. 分析其他币种"
    echo "4. 持续监控模式"
    echo "5. 返回主菜单"
    
    read -p "请选择: " choice
    
    case $choice in
        1)
            ./crypto-analyzer
            ;;
        2)
            ./crypto-analyzer -s ETHUSDT
            ;;
        3)
            read -p "输入币种符号（如SOLUSDT）: " symbol
            ./crypto-analyzer -s $symbol
            ;;
        4)
            read -p "监控间隔（秒，默认300）: " delay
            delay=${delay:-300}
            ./crypto-analyzer -c -d $delay
            ;;
        5)
            return
            ;;
        *)
            echo -e "${RED}无效选择${NC}"
            ;;
    esac
}

# 策略回测菜单
strategy_backtest() {
    echo -e "\n${GREEN}策略回测模块${NC}"
    echo "1. 简单策略（推荐新手）"
    echo "2. 趋势跟踪策略"
    echo "3. 动量突破策略"
    echo "4. 均值回归策略"
    echo "5. 自适应组合策略"
    echo "6. 返回主菜单"
    
    read -p "请选择策略: " choice
    read -p "输入币种（默认BTCUSDT）: " symbol
    symbol=${symbol:-BTCUSDT}
    read -p "回测天数（默认30）: " days
    days=${days:-30}
    
    case $choice in
        1)
            ./backtest -s $symbol -d $days
            ;;
        2)
            ./backtest -s $symbol -d $days --strategy trend
            ;;
        3)
            ./backtest -s $symbol -d $days --strategy momentum
            ;;
        4)
            ./backtest -s $symbol -d $days --strategy reversal
            ;;
        5)
            ./backtest -s $symbol -d $days --strategy combo
            ;;
        6)
            return
            ;;
        *)
            echo -e "${RED}无效选择${NC}"
            ;;
    esac
}

# 双向回测菜单
bidirectional_backtest() {
    echo -e "\n${GREEN}双向交易回测${NC}"
    echo "1. 基础双向策略"
    echo "2. 改进策略（动态止损）"
    echo "3. 仅做多测试"
    echo "4. 自定义参数"
    echo "5. 返回主菜单"
    
    read -p "请选择: " choice
    read -p "输入币种（默认BTCUSDT）: " symbol
    symbol=${symbol:-BTCUSDT}
    read -p "回测天数（默认30）: " days
    days=${days:-30}
    
    case $choice in
        1)
            ./backtest-v2 -s $symbol -d $days
            ;;
        2)
            ./backtest-v2 -s $symbol -d $days --improved
            ;;
        3)
            ./backtest-v2 -s $symbol -d $days --enable-short=false
            ;;
        4)
            read -p "做多阈值（默认0.5）: " long
            long=${long:-0.5}
            read -p "做空阈值（默认-0.5）: " short
            short=${short:--0.5}
            ./backtest-v2 -s $symbol -d $days --long $long --short $short
            ;;
        5)
            return
            ;;
        *)
            echo -e "${RED}无效选择${NC}"
            ;;
    esac
}

# 构建项目
build_project() {
    echo -e "\n${YELLOW}构建项目...${NC}"
    
    echo "构建市场分析器..."
    go build -o crypto-analyzer ./cmd/crypto-analyzer
    
    echo "构建基础回测器..."
    go build -o backtest ./cmd/backtest
    
    echo "构建双向回测器..."
    go build -o backtest-v2 ./cmd/backtest-v2
    
    echo -e "${GREEN}构建完成！${NC}"
}

# 主循环
while true; do
    show_menu
    read -p "请选择功能: " main_choice
    
    case $main_choice in
        1)
            market_analysis
            ;;
        2)
            strategy_backtest
            ;;
        3)
            bidirectional_backtest
            ;;
        4)
            build_project
            ;;
        5)
            echo -e "${GREEN}感谢使用！${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}无效选择，请重试${NC}"
            ;;
    esac
    
    echo -e "\n按回车键继续..."
    read
done