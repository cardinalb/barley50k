#clkm1.3_optimized_loc

library(data.table) # Fast I/O
library(foreach)    # Parallel looping
library(doParallel) # Parallel backend

### Command Line Arguments
args <- commandArgs(trailingOnly = TRUE);

if(length(args) < 4){
  stop("Insufficient arguments provided. Requires: intensityfile thetafile analysis_name workingdir")
}

intensityfile = args[1];
thetafile = args[2];
analysis = args[3];
workingdir = args[4]

### Core Logic Function (kk)
kk = function(theta, intens, cut.12=55, cut.23=45, mindist=0.075,  
              cut.12.IN=65, cut.23.IN=65, cut.NuLL=0.09,
              cut.mean.NuLL=0.20, cut.range.NuLL = 0.50, cut.pcNuLLs=75,
              calling="germplasm", transformation="none", seed=13579) {
  
  i.ok = which(!is.na(theta) & !is.na(intens))
  THeta = theta[i.ok];  INtens = intens[i.ok]
  n.geno = length(THeta)
  n.total = length(theta)
  
  i.EXP = 1:n.geno
  n.NuLL = 0
  i.NuLL = numeric(0)
  
  clstats = rep(0, times=32)
  clstats[c(4,5,6,8:19)] = NA
  clids = rep(NA, times=n.total)
  CLids = rep(NA, times=n.geno)
  
  set.seed(seed)
  q.INtens = quantile(INtens, c(.05,.50,.95))
  low.responder = q.INtens[3] < cut.mean.NuLL
  wide.range = q.INtens[3] - q.INtens[1] > cut.range.NuLL
  clustering.INtens = FALSE
  n.IN.clust = 0
  min.mean.IN = 0
  
  if (wide.range) { 
    clustering.INtens = TRUE
    g.list = lapply(1:5, function(k) kmeans(INtens, k, nstart=5))
    tot.ss = sapply(g.list, function(x) x$tot.withinss)
    gpc.IN = -diff(tot.ss)/tot.ss[1:4] * 100
    gpc.IN = round(gpc.IN, digits=1)
    centers.list = lapply(g.list, function(x) x$centers)
    
    n.IN.clust = 1
    if(gpc.IN[1] > cut.12.IN) n.IN.clust = 2
    if((n.IN.clust == 2) & (gpc.IN[2] > cut.23.IN)) n.IN.clust = 3
    if((n.IN.clust == 3) & (gpc.IN[3] > cut.23.IN)) n.IN.clust = 4
    
    curr.centers = centers.list[[n.IN.clust]]
    min.mean.IN = min(curr.centers)
    i.NuLL.clust = which.min(curr.centers)
    
    if (min.mean.IN < cut.mean.NuLL) { 
      i.NuLL = which(g.list[[n.IN.clust]]$cluster == i.NuLL.clust)
      n.NuLL = length(i.NuLL)
      n.IN.clust = 1
      i.EXP = setdiff(1:n.geno, i.NuLL)
    }
  } else {
    if ((min(INtens) < cut.NuLL) & !(low.responder)) {
      i.NuLL = which(INtens < cut.NuLL)
      n.NuLL = length(i.NuLL)
      n.IN.clust = 1
      i.EXP = setdiff(1:n.geno, i.NuLL)
    }
  }
  
  clstats[21] = round(q.INtens[3], digits=3)
  clstats[22] = round((q.INtens[3] - q.INtens[1]), digits=3)
  clstats[23:32] = 0
  if (clustering.INtens) { 
    locs4 = 29:(29+n.IN.clust-1)
    clstats[23] = n.IN.clust
    clstats[24] = round(min.mean.IN, digits=2)
    clstats[25:28] = round(gpc.IN, digits=2)
    clstats[locs4] = round(as.numeric(centers.list[[n.IN.clust]]), digits=2)
  }
  
  if (n.NuLL > (cut.pcNuLLs * n.geno / 100)) {
    clstats[1] = n.geno
    clstats[3] = n.NuLL
    clstats[7] = round((n.NuLL/n.geno), digits=2) 
    clstats[20] = 1 
    clids[i.ok[i.NuLL]] = ' N'
  } else {
    if (n.NuLL > 0) THeta = THeta[i.EXP]
    
    if (transformation == 'logit') THeta = logit(THeta)
    if (transformation == 'ang')   THeta = asin(sqrt(THeta))
    
    g.list = lapply(1:5, function(k) kmeans(THeta, k, nstart=25)) 
    tot.ss = sapply(g.list, function(x) x$tot.withinss)
    gpc = -diff(tot.ss)/tot.ss[1:4] * 100
    gpc = round(gpc, digits=1)
    g.sizes = lapply(g.list, function(x) x$size)
    g.sds = lapply(1:5, function(i) { sqrt(g.list[[i]]$withinss / pmax(1, g.sizes[[i]] - 1)) })
    g.centers = lapply(g.list, function(x) x$centers)
    
    n.clust = 1
    c2 = g.centers[[2]]; c3 = g.centers[[3]]
    if(gpc[1] > cut.12) {
      d12 = abs(c3[1]-c3[2]); d13 = abs(c3[1]-c3[3]); d23 = abs(c3[2]-c3[3])
      if((gpc[2] > cut.23) & (min(c(d12,d13,d23)) > mindist)) {
        n.clust = 3 
      } else {
        d12_2 = abs(c2[1]-c2[2])
        if (d12_2 > mindist) n.clust = 2
      }
    }
    
    n.clustNuLL = n.clust + (n.NuLL > 0)
    clstats[1] = n.geno
    clstats[2] = n.clust
    clstats[3] = n.NuLL
    clstats[18:19] = 0
    
    curr.clust = g.list[[n.clust]]
    curr.centers = as.numeric(g.centers[[n.clust]])
    ord = order(curr.centers)
    
    locs1 = 4:(4+n.clust-1)
    locs2 = 8:(8+n.clust-1)
    locs3 = 11:(11+n.clust-1)
    
    clstats[locs1] = round(g.sizes[[n.clust]][ord]/(n.geno-n.NuLL), digits=2)
    clstats[7] = round((n.NuLL/n.geno)*(n.NuLL>0), digits=2)
    clstats[locs2] = round(curr.centers[ord], digits=3)
    clstats[locs3] = round(as.numeric(g.sds[[n.clust]][ord]), digits=3)
    clstats[14:17] = round(gpc, digits=2)
    
    if (n.clust > 1) {
      sdmn = sd(curr.centers)
      scale_fac = if(n.clust==2) 0.71 else 0.55
      clstats[18] = round(sdmn/scale_fac, digits=2)
    }
    lmnsd = log(mean(as.numeric(g.sds[[n.clust]]), na.rm=TRUE))
    clstats[19] = round((-1.5 - lmnsd)/5.5, digits=2)
    clstats[20] = as.numeric( (gpc[3] > cut.12) | ((clstats[18] < 0.12) & (n.clust > 1)) | clstats[19] < 0.25 )  
    
    if (calling == "germplasm") {
      CLlabels1 = if((n.clust==1) & mean(THeta)>0.5) 'BB' else 'AA'
      CLlabels2 = c('AA','BB')
      CLlabels3 = c('AA','AB','BB') 
    } else {
      CLlabels1 = 'AA'
      if (n.clust==1) { 
        mn = mean(THeta)
        if (mn >= 0.35) CLlabels1 = if(mn < 0.65) 'AB' else 'BB'
      }
      CLlabels2 = c("AA","BB")
      if (curr.centers[ord][1] > 0.35) CLlabels2 = c('AB','BB')
      if (curr.centers[ord][2] < 0.65) CLlabels2 = c('AA','AB')
      CLlabels3 = c('AA','AB','BB')
    }
    
    list.CLlabels = list(CLlabels1, CLlabels2, CLlabels3)
    tempids = factor(curr.clust$cluster, levels=ord, labels=list.CLlabels[[n.clust]])
    CLids[i.EXP] = as.character(tempids)
    if (n.NuLL > 0) CLids[i.NuLL] = ' N'
    clids[i.ok] = CLids 
  }
  return(list(stats=clstats, ids=clids))
}

################################################################################
### Main Execution Block
################################################################################

my.cut.12 = 55
my.cut.23 = 45
my.cut.12.IN = 65
my.cut.23.IN = 65
my.mindist = 0.075
my.cut.NuLL = 0.09
my.cut.mean.NuLL = 0.20
my.cut.range.NuLL = 0.50
my.cut.pcNuLLs = 75
my.calling = "germplasm"
my.seed = 13579
my.transformation = "none"

### Data Loading
if (!is.na(workingdir) && nchar(workingdir) > 0) {
  setwd(workingdir)
}

cat("Reading Theta file...\n")
dt.theta = fread(thetafile, header=TRUE, data.table=FALSE)
names.snp = dt.theta[,1]
names.geno = colnames(dt.theta)[-1]
n.snp = nrow(dt.theta)
n.geno = ncol(dt.theta) - 1

cat("Reading Intensity file...\n")
dt.intens = fread(intensityfile, header=TRUE, data.table=FALSE)

mat.theta = as.matrix(dt.theta[, -1])
mat.intens = as.matrix(dt.intens[, -1])
rm(dt.theta, dt.intens)
gc()

# --- MODIFIED: Output path determined by input path ---
output_dir = dirname(thetafile)
file.stats = file.path(output_dir, paste("stats", "csv", sep="."))
file.ids   = file.path(output_dir, paste("ids",   "csv", sep="."))
file.parms = file.path(output_dir, paste("parms", "csv", sep="."))
# ----------------------------------------------------

### Parallel Execution
cat("Setting up parallel cluster...\n")
n.cores <- parallel::detectCores() - 1
cl <- makeCluster(n.cores)
registerDoParallel(cl)

cat(paste("Running analysis on", n.snp, "SNPs using", n.cores, "cores...\n"))
results <- foreach(i = 1:n.snp, .packages=c('MASS')) %dopar% {
  kk(mat.theta[i, ], mat.intens[i, ],
     cut.12=my.cut.12, cut.23=my.cut.23, mindist=my.mindist,
     cut.12.IN=my.cut.12.IN, cut.23.IN=my.cut.NuLL, cut.NuLL=my.cut.NuLL,
     cut.mean.NuLL=my.cut.mean.NuLL, cut.range.NuLL=my.cut.range.NuLL,
     cut.pcNuLLs=my.cut.pcNuLLs, calling=my.calling,
     transformation=my.transformation, seed=my.seed)
}
stopCluster(cl)
cat("Analysis complete. Formatting output...\n")

### Output Formatting
names.stats = c("ngeno","nclusters","nNulls","p1","p2","p3","pNULL","mn1","mn2","mn3",
               "sd1","sd2","sd3","%1/2","%2/3","%3/4","%4/5","Sep","Comp","flag",
               "INq95","INRange","n.IN.cl","IN.MinCl","IN%1/2","IN%2/3","IN%3/4","IN%4/5",
               "INmn1","INmn2","IN.mn3","IN.mn4")

stats.matrix <- do.call(rbind, lapply(results, function(x) x$stats))
df.stats <- as.data.frame(stats.matrix)
colnames(df.stats) <- names.stats
rownames(df.stats) <- names.snp
df.stats$orig.snp.id <- 1:n.snp

ids.matrix <- do.call(rbind, lapply(results, function(x) x$ids))
df.ids <- as.data.frame(ids.matrix)
colnames(df.ids) <- names.geno
rownames(df.ids) <- names.snp
df.ids$orig.snp.id <- 1:n.snp

parm.values = c()
parm.names = c('Clustering','Date & Time','workingdir', 'file.stats', 'file.ids', 
               'intensityfile', 'thetafile', 'cut.12.IN', 'cut.23.IN','mindist',
               'minsize', 'range.adj', 'cut.NuLL', 'cut.mean.NuLL', 'cut.range.NuLL', 
               'cut.pcNuLLs', 'seed', 'calling', 'transformation','blank','blank','blank')

date.time = Sys.time()
parm.values[1] = 'K Means (Optimized)'
parm.values[2] = as.character(date.time)
parm.values[3:7] = c(workingdir, file.stats, file.ids, intensityfile, thetafile)
parm.values[8:17] = as.character(c(my.cut.12.IN, my.cut.23.IN, my.mindist, NA, NA, 
                                   my.cut.NuLL, my.cut.mean.NuLL, my.cut.range.NuLL, 
                                   my.cut.pcNuLLs, my.seed))
parm.values[18:22] = c(my.calling, my.transformation,'*','*','*' )
df.parm = data.frame(parm.names, parm.values)

cat("Writing output files...\n")
options(digits=3)
fwrite(df.parm, file.parms)
df.stats.out <- cbind(Row.Names = rownames(df.stats), df.stats)
fwrite(df.stats.out, file.stats, na='*')
df.ids.out <- cbind(Row.Names = rownames(df.ids), df.ids)
fwrite(df.ids.out, file.ids)

cat("Done.\n")