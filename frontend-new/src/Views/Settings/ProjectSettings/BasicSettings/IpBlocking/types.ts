export class FilterIps {
    block_ips: Array<String> = [];

    constructor(ipList: Array<String>) {
        this.block_ips = ipList? ipList : [];
    }

    setIp = (ip: string) => {
        // Check if valid
        if (!IPREGEXP.test(ip)) {  
            return "Invalid Ip Address";
          }  
        
        if (this.block_ips.find((val) => val===ip)) {
            return "Ip Address already excluded";
        }

        this.block_ips.push(ip);

        return true;
    }

    removeIp = (ipToRemove: String) => {
        this.block_ips = this.block_ips.filter((ip) => ip!==ipToRemove);
    }

    getFilterIpPayload = () => {
        return {block_ips: this.block_ips};
    }

    getFilterIpsByChunks = (chunkSize: number = 6) => {
        const chunks: Array<Array<String>> = [];
        for(let i=0;i<this.block_ips.length;i+=chunkSize) {
            chunks.push(this.block_ips.slice(i, i+chunkSize));
        }
        return chunks;
    }
}

const IPREGEXP = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/